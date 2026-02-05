package java

import (
	"bytes"
	ecg "editorconfig-guesser"
	_ "embed"
	"fmt"
	"path/filepath"
	"text/template"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	globs      = []string{"*.java"}
)

type Format struct {
	surveyor          *ecg.BasicSurveyor
	everyFileSurveyor *ecg.BasicSurveyor
	matches           int
}

func (l *Format) SetBasicSurveyor(af *ecg.BasicSurveyor) {
	l.everyFileSurveyor = af
}

func (l *Format) Init() ([]*ecg.SummaryResult, error) {
	return nil, nil
}

func (l *Format) RunFile(f *ecg.File) ([]*ecg.SummaryResult, error) {
	match := false
	for _, gs := range globs {
		_, fn := filepath.Split(f.Filename)
		if m, err := filepath.Match(gs, fn); err != nil {
			return nil, err
		} else if m {
			match = true
			break
		}
	}
	if !match {
		return nil, nil
	}
	l.matches++
	_, _, _, err := l.surveyor.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("running: %w", err)
	}
	return nil, nil
}

func (l *Format) End() ([]*ecg.SummaryResult, error) {
	if l.matches == 0 {
		return nil, nil
	}
	l.surveyor.Summarize()
	return []*ecg.SummaryResult{
		{
			FileGlobs:  globs,
			Confidence: 1,
			Template:   l,
			Path:       "/",
		},
	}, nil
}

func (l *Format) String() (string, error) {
	b := bytes.NewBuffer(nil)
	t := template.Must(template.New("").Parse(string(ectemplate)))
	var allFiles *ecg.BasicSurveyor
	if l.everyFileSurveyor == nil {
		allFiles = l.surveyor
	} else {
		allFiles = ecg.NewBasicSurveyor()
		if l.everyFileSurveyor.InsertFinalNewline != l.surveyor.InsertFinalNewline {
			allFiles.InsertFinalNewline = l.surveyor.InsertFinalNewline
		}
		if l.everyFileSurveyor.Charset != l.surveyor.Charset {
			allFiles.Charset = l.surveyor.Charset
		}
		if l.everyFileSurveyor.Charsets != l.surveyor.Charsets {
			allFiles.Charsets = l.surveyor.Charsets
		}
		if l.everyFileSurveyor.TrimTrailingWhitespace != l.surveyor.TrimTrailingWhitespace {
			allFiles.TrimTrailingWhitespace = l.surveyor.TrimTrailingWhitespace
		}
		if l.everyFileSurveyor.EndOfLine != l.surveyor.EndOfLine {
			allFiles.EndOfLine = l.surveyor.EndOfLine
		}
		if l.everyFileSurveyor.Files != l.surveyor.Files {
			allFiles.Files = l.surveyor.Files
		}
		if l.everyFileSurveyor.CharacterSets != l.surveyor.CharacterSets {
			allFiles.CharacterSets = l.surveyor.CharacterSets
		}

		if l.everyFileSurveyor.IndentStyle != l.surveyor.IndentStyle {
			allFiles.IndentStyle = l.surveyor.IndentStyle
		}
		if l.everyFileSurveyor.IndentSize != l.surveyor.IndentSize {
			allFiles.IndentSize = l.surveyor.IndentSize
		}
		if l.everyFileSurveyor.MaxLineLength != l.surveyor.MaxLineLength {
			allFiles.MaxLineLength = l.surveyor.MaxLineLength
		}
		if l.everyFileSurveyor.TabWidth != l.surveyor.TabWidth {
			allFiles.TabWidth = l.surveyor.TabWidth
		}
	}
	err := t.Execute(b, allFiles)
	return b.String(), err
}

func init() {
	ecg.Register(func() ecg.FileFormat {
		return ecg.NewContainer("Java", &Format{
			surveyor: ecg.NewBasicSurveyor(),
		})
	})
}

var _ ecg.BasicSurveyorSetter = (*Format)(nil)
