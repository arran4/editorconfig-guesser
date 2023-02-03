package gnumake

import (
	"editorconfig-guesser"
	"editorconfig-guesser/fileformat"
	_ "embed"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
	format     ecg.FileFormat = ecg.NewPresence(
		"GNU Make",
		[]string{
			"Makefile",
			"*.mk",
		},
		ectemplate,
	)
)

func init() {
	fileformat.Register(format)
}
