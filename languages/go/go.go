package gnumake

import (
	"bytes"
	"editorconfig-guesser"
	"editorconfig-guesser/languages"
	_ "embed"
	"fmt"
	"path/filepath"
	"sync"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	language   editorconfig_guesser.Language = &Language{
		name: "Go",
		globs: []string{
			"go.mod",
			"go.sum",
			"*.go",
		},
	}
)

func init() {
	languages.Register(language)
}

type Language struct {
	sync.WaitGroup
	reader  chan *editorconfig_guesser.File
	errors  []error
	summary []*editorconfig_guesser.SummaryResult
	name    string
	globs   []string
}

func (l *Language) Name() string {
	return l.name
}

func (l *Language) Start() chan *editorconfig_guesser.File {
	l.reader = make(chan *editorconfig_guesser.File)
	go l.Run()
	return l.reader
}

func (l *Language) Done() ([]*editorconfig_guesser.SummaryResult, error) {
	l.WaitGroup.Wait()
	return l.summary, l.error()
}

func (l *Language) Run() {
	l.WaitGroup.Add(1)
	defer l.WaitGroup.Done()
	for f := range l.reader {
		if f == nil {
			close(l.reader)
			l.reader = nil
			break
		}
		matched := map[int]struct{}{}
		for gsi, gs := range l.globs {
			if m, err := filepath.Match(gs, f.Filename); err != nil {
				l.errors = append(l.errors, fmt.Errorf("matcher %s: %w", gs, err))
			} else if !m {
				continue
			}
			if _, ok := matched[gsi]; ok {
				continue
			}
			matched[gsi] = struct{}{}
			if len(l.summary) == 0 {
				l.summary = append(l.summary, &editorconfig_guesser.SummaryResult{
					FileGlobs:  []string{gs},
					Confidence: 1,
					Template:   bytes.NewBuffer(ectemplate),
					Path:       "/",
				})
			} else {
				l.summary[0].FileGlobs = append(l.summary[0].FileGlobs, gs)
			}
		}
		break
	}
}

func (l *Language) error() error {
	if len(l.errors) == 0 {
		return nil
	}
	return fmt.Errorf("%s errors: %w", l.Name(), l.errors)
}
