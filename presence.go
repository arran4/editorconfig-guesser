package ecg

import (
	"bytes"
	"fmt"
	"path/filepath"
)

// NewPresence ...
func NewPresence(name string, globs []string, ectemplate []byte) FileFormat {
	return NewContainer(name, &Presence{
		globs:      globs,
		ectemplate: ectemplate,
	})
}

// Presence ...
type Presence struct {
	globs      []string
	ectemplate []byte
	matched    map[int]struct{}
	summary    []*SummaryResult
}

// Init ...
func (l *Presence) Init() ([]*SummaryResult, error) {
	l.matched = map[int]struct{}{}
	return nil, nil
}

// RunFile ...
func (l *Presence) RunFile(f *File) ([]*SummaryResult, error) {
	for gsi, gs := range l.globs {
		_, fn := filepath.Split(f.Filename)
		if m, err := filepath.Match(gs, fn); err != nil {
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
				Template:   ErrorStringerWrapper(bytes.NewBuffer(l.ectemplate)),
				Path:       "/",
			})
		} else {
			l.summary[0].FileGlobs = append(l.summary[0].FileGlobs, gs)
		}
	}
	return nil, nil
}

// ErrorStringerWrapperStruct ...
type ErrorStringerWrapperStruct struct {
	stringer fmt.Stringer
}

// String ...
func (e *ErrorStringerWrapperStruct) String() (string, error) {
	return e.stringer.String(), nil
}

// ErrorStringerWrapper ...
func ErrorStringerWrapper(stringer fmt.Stringer) ErrorStringer {
	return &ErrorStringerWrapperStruct{stringer: stringer}
}

// End ...
func (l *Presence) End() ([]*SummaryResult, error) {
	return l.summary, nil
}
