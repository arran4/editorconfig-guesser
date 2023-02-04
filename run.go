package ecg

import (
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"strings"
)

var (
	//go:embed "root.ectemplate"
	rootectemplate []byte
)

func RunInDir(dir fs.FS, ignore func(path string) bool) (string, error) {
	ff := FileFormats()
	chans := make([]chan *File, 0, len(ff))
	for _, eff := range ff {
		chans = append(chans, eff.Start())
	}
	fn := func(path string, d fs.DirEntry, err error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		if ignore(path) {
			return nil
		}
		f := &File{
			Filename:   path,
			FileOpener: dir,
		}
		log.Printf("%v", f)
		for _, e := range chans {
			e <- f
		}
		return nil
	}
	if err := fs.WalkDir(dir, ".", fn); err != nil {
		return "", fmt.Errorf("walking %s: %w", dir, err)
	}
	for _, e := range chans {
		e <- nil
	}
	template := &strings.Builder{}
	_, _ = template.Write(rootectemplate)
	for _, eff := range ff {
		ss, err := eff.Done()
		if err != nil {
			return "", err
		}
		for i, ess := range ss {
			if i > 0 {
				_, _ = fmt.Fprintln(template)
				_, _ = fmt.Fprintln(template)
			}
			_, _ = fmt.Fprintf(template, "[")
			if len(ess.FileGlobs) > 1 {
				_, _ = fmt.Fprintf(template, "{")
			}
			for gi, eg := range ess.FileGlobs {
				if gi > 0 {
					_, _ = fmt.Fprintf(template, ",")
				}
				_, _ = fmt.Fprint(template, eg)
			}
			if len(ess.FileGlobs) > 1 {
				_, _ = fmt.Fprintf(template, "}")
			}
			_, _ = fmt.Fprintf(template, "]\n")
			_, _ = fmt.Fprintln(template, ess.Template.String())
		}
	}
	return template.String(), nil
}
