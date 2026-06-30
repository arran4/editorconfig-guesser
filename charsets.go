// Package ecg guesses the editorconfig settings of a project.
package ecg

import (
	"fmt"
	"golang.org/x/exp/maps"
		"sort"
	"strings"
)

// CharSetSummary ...
type CharSetSummary struct {
	Latin1     int `value:"latin1"`
	Utf8       int `value:"utf-8"`
	Utf16Be    int `value:"utf-16be"`
	Utf16Le    int `value:"utf-16le"`
	Utf8Bom    int `value:"utf-8-bom"`
	Sets       map[string]int
	OtherTotal int
}

// BestFit ...
func (s *CharSetSummary) BestFit() string {
	ks := maps.Keys(s.Sets)
	sort.Slice(ks, func(i, j int) bool {
		a := ks[i]
		b := ks[j]
		return s.Sets[a] > s.Sets[b]
		})
	if len(ks) > 0 {
		return ks[0]
	}
	return ""
}

// Distribution ...
func (s *CharSetSummary) Distribution(total int) string {
	ks := maps.Keys(s.Sets)
	sort.Slice(ks, func(i, j int) bool {
		a := ks[i]
		b := ks[j]
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
