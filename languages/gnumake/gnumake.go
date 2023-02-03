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
		"GNU Make",
		[]string{
			"Makefile",
			"*.mk",
		},
		ectemplate,
	)
)

func init() {
	languages.Register(language)
}
