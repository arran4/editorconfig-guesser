package ecg

import (
	"bytes"
	"path/filepath"
)

func NewPresence(name string, globs []string, ectemplate []byte) FileFormat {
	return NewContainer(name, &Presence{
		globs:      globs,
		ectemplate: ectemplate,
	})
}

type Presence struct {
	globs      []string
	ectemplate []byte
	matched    map[int]struct{}
	summary    []*SummaryResult
}

func (l *Presence) Init() ([]*SummaryResult, error) {
	l.matched = map[int]struct{}{}
	return nil, nil
}

func (l *Presence) RunFile(f *File) ([]*SummaryResult, error) {
	for gsi, gs := range l.globs {
		if m, err := filepath.Match(gs, f.Filename); err != nil {
			return nil, err
		} else if !m {
			continue
		}
		if _, ok := l.matched[gsi]; ok {
			continue
		}
		l.matched[gsi] = struct{}{}
		if len(l.summary) == 0 {
			l.summary = append(l.summary, &SummaryResult{
				FileGlobs:  []string{gs},
				Confidence: 1,
				Template:   bytes.NewBuffer(l.ectemplate),
				Path:       "/",
			})
		} else {
			l.summary[0].FileGlobs = append(l.summary[0].FileGlobs, gs)
		}
	}
	return nil, nil
}

func (l *Presence) End() ([]*SummaryResult, error) {
	return l.summary, nil
}
