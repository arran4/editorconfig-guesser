package allfiles

import (
	"bytes"
	"editorconfig-guesser"
	_ "embed"
	"fmt"
	"text/template"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
)

type Format struct {
	allFiles *ecg.AllFiles
}

func (l *Format) AllFiles() *ecg.AllFiles {
	return l.allFiles
}

func (l *Format) Init() ([]*ecg.SummaryResult, error) {
	return nil, nil
}

func (l *Format) RunFile(f *ecg.File) ([]*ecg.SummaryResult, error) {
	_, _, _, err := l.allFiles.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("running: %w", err)
	}
	return nil, nil
}

func (l *Format) End() ([]*ecg.SummaryResult, error) {
	l.allFiles.Summarize()
	return []*ecg.SummaryResult{
		{
			FileGlobs:  []string{"*"},
			Confidence: 1,
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

func init() {
	ecg.Register(func() ecg.FileFormat {
		return ecg.NewContainer("Typescript", &Format{
			allFiles: &ecg.AllFiles{
				CharacterSets: &ecg.CharSetSummary{
					Sets: map[string]int{},
				},
			},
		})
	})
}

var _ ecg.AllFilesGetter = (*Format)(nil)