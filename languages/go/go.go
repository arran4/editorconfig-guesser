package gnumake

import (
	"editorconfig-guesser"
	"editorconfig-guesser/generic"
	"editorconfig-guesser/languages"
	_ "embed"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	language   editorconfig_guesser.Language = generic.NewPresence(
		"Go",
		[]string{
			"go.mod",
			"go.sum",
			"*.go",
		},
		ectemplate,
	)
)

func init() {
	languages.Register(language)
}
