package editorconfig_guesser

import "fmt"

type SummaryResult struct {
	FileGlobs  []string
	Confidence float64
	Template   fmt.Stringer
	Path       string
}

type File struct {
	Filename string
}

type Language interface {
	Name() string
	Start() chan *File
	Done() ([]*SummaryResult, error)
}
