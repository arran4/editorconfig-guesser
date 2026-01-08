package cli

import (
	ecg "editorconfig-guesser"
	_ "editorconfig-guesser/fileformats"
	"fmt"
	"github.com/denormal/go-gitignore"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Run is a subcommand `ecguess`
// Flags:
//   saveFlag: -s --save (default: false) Save the file as .editorconfig
//   verboseFlag: -v --verbose (default: false) Logs more than what is required
//   args: ... Directories
func Run(saveFlag bool, verboseFlag bool, args ...string) {
	log.SetFlags(log.Flags() | log.Lshortfile)
	if len(args) == 0 {
		fmt.Println("Please provide at least one directory")
	}
	for _, e := range args {
		ignore, err := gitignore.NewRepository(e)
		if err != nil {
			log.Printf("Loading git ignores failed: %s", err)
			ignore = nil
		}

		template, err := ecg.RunInDir(os.DirFS(e), func(file *ecg.File) bool {
			for _, part := range filepath.SplitList(file.Filename) {
				if strings.HasPrefix(part, ".") && part != "." {
					if verboseFlag {
						log.Printf("Skipping %s as it has a hidden file in the path: %s", file.Filename, part)
					}
					return true
				}
			}
			if ignore != nil && ignore.Ignore(filepath.Join(e, file.Filename)) {
				if verboseFlag {
					log.Printf("Skipping %s as it is in the .gitignore file", file.Filename)
				}
				return true
			}
			if file.IsBinary() {
				if verboseFlag {
					log.Printf("Skipping %s as it is considered a binary file", file.Filename)
				}
				return true
			}
			return false
		})
		if err != nil {
			log.Panicf("Error: %s", err)
		}
		if len(args) > 1 {
			fmt.Println("// ", e)
		}
		fmt.Println(template)
		if len(args) > 1 {
			fmt.Println()
			fmt.Println()
		}
		if saveFlag {
			outfn := filepath.Join(e, ".editorconfig")
			if err := os.WriteFile(outfn, []byte(template), 0644); err != nil {
				log.Panicf("Error saving %s because %s", outfn, err)
			} else {
				log.Println("Wrote: ", outfn)
			}
		}
	}
}
