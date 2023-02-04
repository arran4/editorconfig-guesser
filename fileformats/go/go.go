package gnumake

import (
	"editorconfig-guesser"
	_ "embed"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
)

func init() {
	ecg.Register(func() ecg.FileFormat {
		return ecg.NewPresence(
			"Go",
			[]string{
				"go.mod",
				"go.sum",
				"*.go",
			},
			ectemplate,
		)
	})
}
