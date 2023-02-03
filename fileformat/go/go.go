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
	fileformat.Register(format)
}
