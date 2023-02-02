package editorconfig_guesser

type SummaryResult struct {
	FileGlobs  []string
	Rules      map[string]any
	Confidence float64
}

type File struct {
	Filename string
}

type Language interface {
	Run() chan *File
	Done() (*SummaryResult, error)
}
