package main

import (
	ecg "editorconfig-guesser"
	_ "editorconfig-guesser/fileformats"
	"flag"
	"fmt"
	"github.com/denormal/go-gitignore"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	saveFlag    = flag.Bool("save", false, "Save the file as .editorconfig")
	verboseFlag = flag.Bool("verbose", false, "Logs more than what is required")
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()
	for _, e := range flag.Args() {
		ignore, err := gitignore.NewRepository(e)
		if err != nil {
			log.Panicf("Loading git ignores")
		}

		template, err := ecg.RunInDir(os.DirFS(e), func(file *ecg.File) bool {
			for _, e := range filepath.SplitList(file.Filename) {
				if strings.HasPrefix(e, ".") && e != "." {
					if *verboseFlag {
						log.Printf("Skipping %s as it has a hidden file in the path", e)
					}
					return true
				}
			}
			if ignore.Ignore(file.Filename) {
				if *verboseFlag {
					log.Printf("Skipping %s as it is in the .gitignore file", e)
				}
			}
			if file.IsBinary() {
				if *verboseFlag {
					log.Printf("Skipping %s as it is considered a binary file", e)
				}
			}
			return false
		})
		if err != nil {
			log.Panicf("Error: %s", err)
		}
		if len(flag.Args()) > 1 {
			fmt.Println("// ", e)
		}
		fmt.Println(template)
		if len(flag.Args()) > 1 {
			fmt.Println()
			fmt.Println()
		}
		if *saveFlag {
			outfn := filepath.Join(e, ".editorconfig")
			if err := os.WriteFile(outfn, []byte(template), 0644); err != nil {
				log.Panicf("Error saving %s because %s", outfn, err)
			} else {
				log.Println("Wrote: ", outfn)
			}
		}
	}
}
