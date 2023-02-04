package ecg

import (
	"errors"
	"fmt"
	"github.com/saintfish/chardet"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
)

// Contraster Provides contrasts two SummaryResults from the same file format for the purpose of consolidation -- future -- maybe
type Contraster func(s1, s2 *SummaryResult) int

type ErrorStringer interface {
	String() (string, error)
}

// SummaryResult results from a file format
type SummaryResult struct {
	// Impacted globs, used for reconstructing.
	FileGlobs []string
	// How confident we are not used atm
	Confidence float64
	// Future use to compare how different two file paths (same file format) are
	Contaster Contraster
	// The template
	Template ErrorStringer
	// The path, this is for future versions where it will suggest sub-directory variants based on confidence and contrast -- maybe
	Path string
	// Internal data, probably going to be used by Contraster
	Data any
}

// File reference, could also be a cache
type File struct {
	Filename   string
	FileOpener fs.FS
	size       *int64
	sync.Mutex
}

// Size of file + cache
func (fd *File) Size() int64 {
	fd.Lock()
	defer fd.Unlock()
	if fd.size != nil {
		return *fd.size
	}
	st, err := os.Stat(fd.Filename)
	if err != nil {
		return -1
	}
	s := st.Size()
	fd.size = &s
	return *fd.size
}

// Open abstracter eventually might cache, perhaps checking file size first - or only caching the first 256kb
func (fd *File) Open() (io.ReadSeekCloser, error) {
	if fd.FileOpener != nil {
		f, err := fd.FileOpener.Open(fd.Filename)
		rsc, ok := f.(io.ReadSeekCloser)
		if !ok {
			return nil, fmt.Errorf("file isn't readable, seakable or closable")
		}
		return rsc, err
	}
	return os.Open(fd.Filename)
}

// FileFormat a file format
type FileFormat interface {
	// Name The display name for errors etc
	Name() string
	// Start starts reading files sent to it on the channel, will close on receiving a nil
	Start() chan *File
	// Done waits until Start() is complete, then returns the SummaryResults and/or an error
	Done() ([]*SummaryResult, error)
}

type AllFiles struct {
	InsertFinalNewline     string
	Charset                string
	Charsets               string
	TrimTrailingWhitespace string
	EndOfLine              string
	Files                  int
	CharacterSets          *CharSetSummary
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

func (l *AllFiles) ReadFile(fd *File) (string, bool, *LineSurvey, error) {
	l.Files += 1
	f, err := fd.Open()
	if err != nil {
		return "", false, nil, fmt.Errorf("opening %s: %w", fd.Filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error: closing %s: %s", fd.Filename, err)
		}
	}()
	b := make([]byte, ReadSize)
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
	survey := LineSurveySample(b)
	if n, err := f.Seek(-1, io.SeekEnd); err != nil {
		return "", false, nil, fmt.Errorf("seak %d from %s: %w", n, fd.Filename, err)
	} else if n > 0 {
		b = make([]byte, 1)
		if n, err := f.Read(b); err != nil {
			return "", false, nil, fmt.Errorf("read %d (of %d) from %s: %w", n, len(b), fd.Filename, err)
		}
		finalNewLine = b[0] == '\n'
	}
	switch charset {
	case "UTF-8":
		l.CharacterSets.Utf8 += 1
	case "UTF-16BE":
		l.CharacterSets.Utf16Be += 1
	case "UTF-16LE":
		l.CharacterSets.Utf16Le += 1
	case "ISO-8859-1":
		l.CharacterSets.Latin1 += 1
	case "":
	default:
		l.CharacterSets.OtherTotal += 1
	}
	l.CharacterSets.Sets[charset] += 1
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
	return charset, finalNewLine, survey, nil
}

func (l *AllFiles) Summarize() {
	if l.Files > 0 && (l.finalNewLineBalance.True*100)/l.Files > 80 { // 20% threshold
		l.InsertFinalNewline = "true"
	} else if l.Files > 0 && (l.finalNewLineBalance.False*100)/l.Files > 80 { // 20% threshold
		l.InsertFinalNewline = "false"
	}
	l.Charset = l.CharacterSets.BestFit()
	l.Charsets = l.CharacterSets.Distribution(l.Files)
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
}

type AllFilesGetter interface {
	AllFiles() *AllFiles
}

type AllFilesSetter interface {
	AllFiles(af *AllFiles)
}
