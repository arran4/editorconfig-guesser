package gnumake

import (
	"editorconfig-guesser"
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
	ecg.Register(format)
}
