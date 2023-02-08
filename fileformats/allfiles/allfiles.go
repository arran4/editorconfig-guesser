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
	allFiles *ecg.BasicSurveyor
}

func (l *Format) BasicSurveyor() *ecg.BasicSurveyor {
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
			Confidence: 3,
			Template:   l,
			Path:       "/",
		},
	}, nil
}

func (l *Format) String() (string, error) {
	b := bytes.NewBuffer(nil)
	t := template.Must(template.New("").Parse(string(ectemplate)))
	err := t.Execute(b, l.allFiles)
	return b.String(), err
}

func init() {
	ecg.Register(func() ecg.FileFormat {
		return ecg.NewContainer("ALl Files", &Format{
			allFiles: ecg.NewBasicSurveyor(),
		})
	})
}

var _ ecg.BasicSurveyorGetter = (*Format)(nil)
