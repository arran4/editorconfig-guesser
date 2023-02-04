package allfiles

import (
	"bytes"
	"editorconfig-guesser"
	_ "embed"
	"errors"
	"fmt"
	"github.com/saintfish/chardet"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"log"
	"strings"
	"sync"
	"text/template"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	format     = &Format{}
)

type Format struct {
	sync.WaitGroup
	reader                 chan *ecg.File
	errors                 []error
	ectemplate             []byte
	InsertFinalNewline     string
	Charset                string
	Charsets               string
	TrimTrailingWhitespace string
	EndOfLine              string
}

func (l *Format) String() string {
	b := bytes.NewBuffer(nil)
	t := template.Must(template.New("").Parse(string(ectemplate)))
	_ = t.Execute(b, l)
	return b.String()
}

func (l *Format) Name() string {
	return "ALl Files"
}

func (l *Format) Start() chan *ecg.File {
	l.reader = make(chan *ecg.File)
	l.WaitGroup.Add(1)
	go l.Run()
	return l.reader
}

func (l *Format) Done() ([]*ecg.SummaryResult, error) {
	l.WaitGroup.Wait()
	return []*ecg.SummaryResult{
		{
			FileGlobs:  []string{"*"},
			Confidence: 3,
			Template:   l,
			Path:       "/",
		},
	}, l.error()
}

type CharSetSummary struct {
	Latin1     int `value:"latin1"`
	Utf8       int `value:"utf-8"`
	Utf16Be    int `value:"utf-16be"`
	Utf16Le    int `value:"utf-16le"`
	Utf8Bom    int `value:"utf-8-bom"`
	Sets       map[string]int
	OtherTotal int
}

func (s *CharSetSummary) BestFit() string {
	ks := maps.Keys(s.Sets)
	slices.SortFunc(ks, func(a, b string) bool {
		return s.Sets[a] > s.Sets[b]
	})
	if len(ks) > 0 {
		return ks[0]
	}
	return ""
}

func (s *CharSetSummary) Distribution(total int) string {
	ks := maps.Keys(s.Sets)
	slices.SortFunc(ks, func(a, b string) bool {
		return s.Sets[a] > s.Sets[b]
	})
	r := &strings.Builder{}
	for i, e := range ks {
		if i > 0 {
			r.WriteString(", ")
		}
		_, _ = fmt.Fprintf(r, "%s (%0.1f%%)", e, float64(s.Sets[e])/float64(total)*100)
	}
	return r.String()
}

func (l *Format) Run() {
	defer l.WaitGroup.Done()
	var Files int
	characterSets := &CharSetSummary{
		Sets: map[string]int{},
	}
	finalNewLineBalance := struct {
		True  int
		False int
	}{}
	trailingSpaceOkay := struct {
		True  int
		False int
	}{}
	lineEndings := struct {
		Windows int
		Unix    int
	}{}
	for f := range l.reader {
		if f == nil {
			close(l.reader)
			l.reader = nil
			break
		}
		Files += 1
		charset, finalNewLine, survey, err := l.runOnFile(f)
		if err != nil {
			l.errors = append(l.errors, fmt.Errorf("running: %w", err))
			continue
		}
		switch charset {
		case "UTF-8":
			characterSets.Utf8 += 1
		case "UTF-16BE":
			characterSets.Utf16Be += 1
		case "UTF-16LE":
			characterSets.Utf16Le += 1
		case "ISO-8859-1":
			characterSets.Latin1 += 1
		case "":
		default:
			characterSets.OtherTotal += 1
		}
		characterSets.Sets[charset] += 1
		if finalNewLine {
			finalNewLineBalance.True += 1
		} else {
			finalNewLineBalance.False += 1
		}
		if survey.WindowNewlines < survey.NewLines/5 { // 20% threshold
			lineEndings.Unix += 1
		} else if survey.WindowNewlines > survey.NewLines/5*4 { // 80% threshold
			lineEndings.Windows += 1
		}
		if !survey.TrailingWhitespaceCommon() {
			trailingSpaceOkay.True += 1
		} else {
			trailingSpaceOkay.False += 1
		}
	}
	if Files > 0 && (finalNewLineBalance.True*100)/Files > 80 { // 20% threshold
		l.InsertFinalNewline = "true"
	} else if Files > 0 && (finalNewLineBalance.False*100)/Files > 80 { // 20% threshold
		l.InsertFinalNewline = "false"
	}
	l.Charset = characterSets.BestFit()
	l.Charsets = characterSets.Distribution(Files)
	if Files > 0 && (trailingSpaceOkay.True*100)/Files > 80 { // 20% threshold
		l.TrimTrailingWhitespace = "true"
	} else if Files > 0 && (trailingSpaceOkay.False*100)/Files > 80 { // 20% threshold
		l.TrimTrailingWhitespace = "false"
	}
	if Files > 0 && (lineEndings.Unix*100)/Files > 80 { // 20% threshold
		l.EndOfLine = "lf"
	} else if Files > 0 && (lineEndings.Windows*100)/Files > 80 { // 80% threshold
		l.EndOfLine = "crlf"
	}
}

func (l *Format) error() error {
	if len(l.errors) == 0 {
		return nil
	}
	return fmt.Errorf("%s errors: %w", l.Name(), l.errors)
}

func (l *Format) runOnFile(fd *ecg.File) (string, bool, *ecg.LineSurvey, error) {
	f, err := fd.Open()
	if err != nil {
		return "", false, nil, fmt.Errorf("opening %s: %w", fd.Filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error: closing %s: %s", fd.Filename, err)
		}
	}()
	b := make([]byte, ecg.ReadSize)
	if n, err := f.Read(b); err != nil {
		return "", false, nil, fmt.Errorf("read %d (of %d) from %s: %w", n, len(b), fd.Filename, err)
	}
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(b)
	if err != nil {
		if !errors.Is(err, chardet.NotDetectedError) {
			return "", false, nil, fmt.Errorf("detect character encoding from %s: %w", fd.Filename, err)
		}
	}
	var charset string
	if result != nil && result.Confidence > 80 {
		// TODO use with Summary's Confidence
		charset = result.Charset
	}
	var finalNewLine bool
	sample := ecg.LineSurveySample(b)
	if n, err := f.Seek(-1, io.SeekEnd); err != nil {
		return "", false, nil, fmt.Errorf("seak %d from %s: %w", n, fd.Filename, err)
	} else if n > 0 {
		b = make([]byte, 1)
		if n, err := f.Read(b); err != nil {
			return "", false, nil, fmt.Errorf("read %d (of %d) from %s: %w", n, len(b), fd.Filename, err)
		}
		finalNewLine = b[0] == '\n'
	}

	return charset, finalNewLine, sample, nil
}

func init() {
	ecg.Register(format)
}
