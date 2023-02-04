package ecg

import (
	"sort"
	"strings"
)

var (
	fileFormats []FileFormat
	sorted      = false
)

func Register(fileFormat FileFormat) {
	fileFormats = append(fileFormats, fileFormat)
}

func FileFormats() []FileFormat {
	if !sorted {
		sort.Sort(FileFormatsSorter(fileFormats))
		sorted = true
	}
	return fileFormats
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
