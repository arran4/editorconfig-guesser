package ecg

import (
	"fmt"
	"io"
	"os"
)

// Contraster Provides contrasts two SummaryResults from the same file format for the purpose of consolidation -- future -- maybe
type Contraster func(s1, s2 *SummaryResult) int

// SummaryResult results from a file format
type SummaryResult struct {
	// Impacted globs, used for reconstructing.
	FileGlobs []string
	// How confident we are not used atm
	Confidence float64
	// Future use to compare how different two file paths (same file format) are
	Contaster Contraster
	// The template
	Template fmt.Stringer
	// The path, this is for future versions where it will suggest sub-directory variants based on confidence and contrast -- maybe
	Path string
	// Internal data, probably going to be used by Contraster
	Data any
}

// File reference, could also be a cache
type File struct {
	Filename string
}

// Open abstracter eventually might cache, perhaps checking file size first - or only caching the first 256kb
func (f File) Open() (io.ReadCloser, error) {
	return os.Open(f.Filename)
}

// FileFormat a file format
type FileFormat interface {
	// Name The display name for errors etc
	Name() string
	// Start starts reading files sent to it on the channel, will close on receiving a nil
	Start() chan *File
	// Done waits until Start() is complete, then returns the SummaryResults and/or an error
	Done() ([]*SummaryResult, error)
}
