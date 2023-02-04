package ecg

import (
	"fmt"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"
)

type CharSetSummary struct {
	Latin1     int `value:"latin1"`
	Utf8       int `value:"utf-8"`
	Utf16Be    int `value:"utf-16be"`
	Utf16Le    int `value:"utf-16le"`
	Utf8Bom    int `value:"utf-8-bom"`
	Sets       map[string]int
	OtherTotal int
}

func (s *CharSetSummary) BestFit() string {
	ks := maps.Keys(s.Sets)
	slices.SortFunc(ks, func(a, b string) bool {
		return s.Sets[a] > s.Sets[b]
	})
	if len(ks) > 0 {
		return ks[0]
	}
	return ""
}

func (s *CharSetSummary) Distribution(total int) string {
	ks := maps.Keys(s.Sets)
	slices.SortFunc(ks, func(a, b string) bool {
		return s.Sets[a] > s.Sets[b]
	})
	r := &strings.Builder{}
	for i, e := range ks {
		if i > 0 {
			r.WriteString(", ")
		}
		_, _ = fmt.Fprintf(r, "%s (%0.1f%%)", e, float64(s.Sets[e])/float64(total)*100)
	}
	return r.String()
}
