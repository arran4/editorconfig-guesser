package languages

import (
	editorconfig_guesser "editorconfig-guesser"
	"sort"
	"strings"
)

var (
	languages []editorconfig_guesser.Language
	sorted    = false
)

func Register(language editorconfig_guesser.Language) {
	languages = append(languages, language)
}

func Languages() []editorconfig_guesser.Language {
	if !sorted {
		sort.Sort(LanguageSorter(languages))
		sorted = true
	}
	return languages
}

type LanguageSorter []editorconfig_guesser.Language

func (l LanguageSorter) Len() int {
	return len(l)
}

func (l LanguageSorter) Less(i, j int) bool {
	// I actually don't care about the order right now, just needs to be consistent
	return strings.Compare(l[i].Name(), l[j].Name()) > 0
}

func (l LanguageSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
