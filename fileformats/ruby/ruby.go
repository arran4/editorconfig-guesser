package ruby

import (
	ecg "editorconfig-guesser"
	_ "embed"
)

var (
	//go:embed "ectemplate"
	ectemplate []byte
)

func init() {
	ecg.Register(func() ecg.FileFormat {
		return ecg.NewPresence(
			"Ruby",
			[]string{
				"*.rb",
				"Rakefile",
				"Gemfile",
			},
			ectemplate,
		)
	})
}
