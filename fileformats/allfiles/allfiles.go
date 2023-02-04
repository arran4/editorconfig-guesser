package allfiles

import (
	"bytes"
	"editorconfig-guesser"
	_ "embed"
	"errors"
	"fmt"
	"github.com/saintfish/chardet"
	"io"
	"log"
	"text/template"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
)

type Format struct {
	InsertFinalNewline     string
	Charset                string
	Charsets               string
	TrimTrailingWhitespace string
	EndOfLine              string
	Files                  int
	characterSets          *ecg.CharSetSummary
	finalNewLineBalance    struct {
		True  int
		False int
	}
	trailingSpaceOkay struct {
		True  int
		False int
	}
	lineEndings struct {
		Windows int
		Unix    int
	}
}

func (l *Format) Init() ([]*ecg.SummaryResult, error) {
	l.characterSets = &ecg.CharSetSummary{
		Sets: map[string]int{},
	}
	return nil, nil
}

func (l *Format) RunFile(f *ecg.File) ([]*ecg.SummaryResult, error) {
	l.Files += 1
	charset, finalNewLine, survey, err := l.readFile(f)
	if err != nil {
		return nil, fmt.Errorf("running: %w", err)
	}
	switch charset {
	case "UTF-8":
		l.characterSets.Utf8 += 1
	case "UTF-16BE":
		l.characterSets.Utf16Be += 1
	case "UTF-16LE":
		l.characterSets.Utf16Le += 1
	case "ISO-8859-1":
		l.characterSets.Latin1 += 1
	case "":
	default:
		l.characterSets.OtherTotal += 1
	}
	l.characterSets.Sets[charset] += 1
	if finalNewLine {
		l.finalNewLineBalance.True += 1
	} else {
		l.finalNewLineBalance.False += 1
	}
	if survey.WindowNewlines < survey.NewLines/5 { // 20% threshold
		l.lineEndings.Unix += 1
	} else if survey.WindowNewlines > survey.NewLines/5*4 { // 80% threshold
		l.lineEndings.Windows += 1
	}
	if !survey.TrailingWhitespaceCommon() {
		l.trailingSpaceOkay.True += 1
	} else {
		l.trailingSpaceOkay.False += 1
	}
	return nil, nil
}

func (l *Format) End() ([]*ecg.SummaryResult, error) {
	if l.Files > 0 && (l.finalNewLineBalance.True*100)/l.Files > 80 { // 20% threshold
		l.InsertFinalNewline = "true"
	} else if l.Files > 0 && (l.finalNewLineBalance.False*100)/l.Files > 80 { // 20% threshold
		l.InsertFinalNewline = "false"
	}
	l.Charset = l.characterSets.BestFit()
	l.Charsets = l.characterSets.Distribution(l.Files)
	if l.Files > 0 && (l.trailingSpaceOkay.True*100)/l.Files > 80 { // 20% threshold
		l.TrimTrailingWhitespace = "true"
	} else if l.Files > 0 && (l.trailingSpaceOkay.False*100)/l.Files > 80 { // 20% threshold
		l.TrimTrailingWhitespace = "false"
	}
	if l.Files > 0 && (l.lineEndings.Unix*100)/l.Files > 80 { // 20% threshold
		l.EndOfLine = "lf"
	} else if l.Files > 0 && (l.lineEndings.Windows*100)/l.Files > 80 { // 80% threshold
		l.EndOfLine = "crlf"
	}
	return []*ecg.SummaryResult{
		{
			FileGlobs:  []string{"*"},
			Confidence: 3,
			Template:   l,
			Path:       "/",
		},
	}, nil
}

func (l *Format) String() string {
	b := bytes.NewBuffer(nil)
	t := template.Must(template.New("").Parse(string(ectemplate)))
	_ = t.Execute(b, l)
	return b.String()
}

func (l *Format) readFile(fd *ecg.File) (string, bool, *ecg.LineSurvey, error) {
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
	ecg.Register(func() ecg.FileFormat {
		return ecg.NewContainer("ALl Files", &Format{})
	})
}
