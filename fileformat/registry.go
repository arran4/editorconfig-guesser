package fileformat

import (
	"editorconfig-guesser"
	"sort"
	"strings"
)

var (
	fileFormats []ecg.FileFormat
	sorted      = false
)

func Register(fileFormat ecg.FileFormat) {
	fileFormats = append(fileFormats, fileFormat)
}

func FileFormats() []ecg.FileFormat {
	if !sorted {
		sort.Sort(FileFormatsSorter(fileFormats))
		sorted = true
	}
	return fileFormats
}

type FileFormatsSorter []ecg.FileFormat

func (l FileFormatsSorter) Len() int {
	return len(l)
}

func (l FileFormatsSorter) Less(i, j int) bool {
	// I actually don't care about the order right now, just needs to be consistent
	return strings.Compare(l[i].Name(), l[j].Name()) > 0
}

func (l FileFormatsSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
