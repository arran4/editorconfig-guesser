package allfiles

import (
	"bytes"
	"editorconfig-guesser"
	_ "embed"
	"fmt"
	"path/filepath"
	"text/template"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	globs      = []string{"*.ts"}
)

type Format struct {
	allFiles      *ecg.AllFiles
	actualAllFile *ecg.AllFiles
	matches       int
}

func (l *Format) SetAllFiles(af *ecg.AllFiles) {
	l.actualAllFile = af
}

func (l *Format) AllFiles() *ecg.AllFiles {
	return l.allFiles
}

func (l *Format) Init() ([]*ecg.SummaryResult, error) {
	return nil, nil
}

func (l *Format) RunFile(f *ecg.File) ([]*ecg.SummaryResult, error) {
	for _, gs := range globs {
		if m, err := filepath.Match(gs, f.Filename); err != nil {
			return nil, err
		} else if !m {
			return nil, nil
		}
	}
	l.matches++
	_, _, _, err := l.allFiles.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("running: %w", err)
	}
	return nil, nil
}

func (l *Format) End() ([]*ecg.SummaryResult, error) {
	if l.matches == 0 {
		return nil, nil
	}
	l.allFiles.Summarize()
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
	var allFiles *ecg.AllFiles
	if l.actualAllFile == nil {
		allFiles = l.allFiles
	} else {
		allFiles = &ecg.AllFiles{}
		if l.actualAllFile.InsertFinalNewline != l.allFiles.InsertFinalNewline {
			allFiles.InsertFinalNewline = l.allFiles.InsertFinalNewline
		}
		if l.actualAllFile.Charset != l.allFiles.Charset {
			allFiles.Charset = l.allFiles.Charset
		}
		if l.actualAllFile.Charsets != l.allFiles.Charsets {
			allFiles.Charsets = l.allFiles.Charsets
		}
		if l.actualAllFile.TrimTrailingWhitespace != l.allFiles.TrimTrailingWhitespace {
			allFiles.TrimTrailingWhitespace = l.allFiles.TrimTrailingWhitespace
		}
		if l.actualAllFile.EndOfLine != l.allFiles.EndOfLine {
			allFiles.EndOfLine = l.allFiles.EndOfLine
		}
		if l.actualAllFile.Files != l.allFiles.Files {
			allFiles.Files = l.allFiles.Files
		}
		if l.actualAllFile.CharacterSets != l.allFiles.CharacterSets {
			allFiles.CharacterSets = l.allFiles.CharacterSets
		}
	}
	err := t.Execute(b, allFiles)
	return b.String(), err
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
var _ ecg.AllFilesSetter = (*Format)(nil)
