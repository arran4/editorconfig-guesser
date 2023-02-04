package ecg

import (
	"fmt"
	"sync"
)

type FileRunner interface {
	Init() ([]*SummaryResult, error)
	RunFile(f *File) ([]*SummaryResult, error)
	End() ([]*SummaryResult, error)
}

func NewContainer(name string, fr FileRunner) *Container {
	return &Container{
		name:       name,
		FileRunner: fr,
	}
}

type Container struct {
	sync.WaitGroup
	reader  chan *File
	errors  []error
	summary []*SummaryResult
	name    string
	FileRunner
}

func (l *Container) Name() string {
	return l.name
}

func (l *Container) Start() chan *File {
	l.reader = make(chan *File)
	l.WaitGroup.Add(1)
	go l.Run()
	return l.reader
}

func (l *Container) Done() ([]*SummaryResult, error) {
	l.WaitGroup.Wait()
	return l.summary, l.error()
}

func (l *Container) Run() {
	defer l.WaitGroup.Done()
	if sr, err := l.FileRunner.Init(); err != nil {
		l.errors = append(l.errors, err)
	} else if len(sr) > 0 {
		l.summary = append(l.summary, sr...)
	}
	for f := range l.reader {
		if f == nil {
			close(l.reader)
			l.reader = nil
			break
		}
		if sr, err := l.FileRunner.RunFile(f); err != nil {
			l.errors = append(l.errors, err)
		} else if len(sr) > 0 {
			l.summary = append(l.summary, sr...)
		}
	}
	if sr, err := l.FileRunner.End(); err != nil {
		l.errors = append(l.errors, err)
	} else if len(sr) > 0 {
		l.summary = append(l.summary, sr...)
	}
}

func (l *Container) error() error {
	if len(l.errors) == 0 {
		return nil
	}
	return fmt.Errorf("%s errors: %w", l.Name(), l.errors[0])
}
