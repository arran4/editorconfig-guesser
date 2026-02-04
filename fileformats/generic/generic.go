package generic

import (
	"bytes"
	"editorconfig-guesser"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	globs      = [][]string{
		[]string{
			"*.ts", // TODO: Add full support
			"*.js", // TODO: Add full support
		},
		[]string{
			"*.cpp", // TODO: Add full support
			"*.h",   // TODO: Add full support
			"*.c",   // TODO: Add full support
		},
		[]string{
			"*.cs", // TODO: Add full support
		},
		[]string{
			"*.json", // TODO: Add full support
		},
		[]string{
			"*.yaml", // TODO: Add full support
			"*.yml",  // TODO: Add full support
		},
		[]string{
			"*.xml", // TODO: Add full support
		},
		[]string{
			"*.html", // TODO: Add full support
			"*.htm",  // TODO: Add full support
		},
		[]string{
			"*.css", // TODO: Add full support
		},
		[]string{
			"*.php", // TODO: Add full support
		},
		[]string{
			"*.md", // TODO: Add full support
		},
		[]string{
			"*.sh", // TODO: Add full support
		},
	}
)

type Format struct {
	surveyor          map[string]*ecg.BasicSurveyor
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
	var match []string
	for _, gs := range globs {
		_, fn := filepath.Split(f.Filename)
		for _, gss := range gs {
			if m, err := filepath.Match(gss, fn); err != nil {
				return nil, err
			} else if m {
				match = gs
				break
			}
		}
		if len(match) > 0 {
			break
		}
	}
	if len(match) == 0 {
		return nil, nil
	}
	l.matches++
	globstr := strings.Join(match, ":")
	surveyor, ok := l.surveyor[globstr]
	if !ok {
		surveyor = ecg.NewBasicSurveyor()
		l.surveyor[globstr] = surveyor
	}
	_, _, _, err := surveyor.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("running: %w", err)
	}
	return nil, nil
}

func (l *Format) End() ([]*ecg.SummaryResult, error) {
	if l.matches == 0 {
		return nil, nil
	}
	results := make([]*ecg.SummaryResult, 0, len(globs))
	for _, gs := range globs {
		surveyor, ok := l.surveyor[strings.Join(gs, ":")]
		if !ok {
			continue
		}
		surveyor.Summarize()
		results = append(results, &ecg.SummaryResult{
			FileGlobs:  gs,
			Confidence: 1,
			Template: &Surveyor{
				everyFileSurveyor: l.everyFileSurveyor,
				surveyor:          surveyor,
			},
			Path: "/",
		})
	}
	return results, nil
}

type Surveyor struct {
	everyFileSurveyor *ecg.BasicSurveyor
	surveyor          *ecg.BasicSurveyor
}

func (l *Surveyor) String() (string, error) {
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
		return ecg.NewContainer("Generic", &Format{
			surveyor: map[string]*ecg.BasicSurveyor{},
		})
	})
}

var _ ecg.BasicSurveyorSetter = (*Format)(nil)
