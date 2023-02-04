package ecg

import (
	"sort"
	"strings"
)

type FileFormatFactory func() FileFormat

var (
	fileFormats []FileFormatFactory
	sorted      = false
)

func Register(fileFormat FileFormatFactory) {
	fileFormats = append(fileFormats, fileFormat)
}

func FileFormats() []FileFormat {
	ffs := make([]FileFormat, len(fileFormats), len(fileFormats))
	var af *AllFiles
	for i, fff := range fileFormats {
		ffs[i] = fff()
		if afg, ok := ffs[i].(AllFilesGetter); ok {
			af = afg.AllFiles()
		}
	}
	if af != nil {
		for _, ff := range ffs {
			if afs, ok := ff.(AllFilesSetter); ok {
				afs.AllFiles(af)
			}
		}
	}
	if !sorted {
		sort.Sort(FileFormatsSorter(ffs))
		sorted = true
	}
	return ffs
}

type FileFormatsSorter []FileFormat

func (l FileFormatsSorter) Len() int {
	return len(l)
}

func (l FileFormatsSorter) Less(i, j int) bool {
	// I actually don't care about the order right now, just needs to be consistent and All Files first.
	// TODO add priority perhaps dynamic based on confidence?
	return strings.Compare(l[i].Name(), l[j].Name()) < 0
}

func (l FileFormatsSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
