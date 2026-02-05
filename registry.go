package ecg

import (
	"sort"
	"strings"
)

type FileFormatFactory func() FileFormat

var (
	fileFormats []FileFormatFactory
)

func Register(fileFormat FileFormatFactory) {
	fileFormats = append(fileFormats, fileFormat)
}

func FileFormats() []FileFormat {
	ffs := make([]FileFormat, len(fileFormats))
	var af *BasicSurveyor
	for i, fff := range fileFormats {
		ffs[i] = fff()
		if afg, ok := ffs[i].(BasicSurveyorGetter); ok {
			af = afg.BasicSurveyor()
		}
	}
	if af != nil {
		for _, ff := range ffs {
			if afs, ok := ff.(BasicSurveyorSetter); ok {
				afs.SetBasicSurveyor(af)
			}
		}
	}
	sort.Sort(FileFormatsSorter(ffs))
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
