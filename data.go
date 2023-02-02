package editorconfig_guesser

import "embed"

var (
	//go:embed "languages/**/ectemplate"
	languagesFs embed.FS
)
