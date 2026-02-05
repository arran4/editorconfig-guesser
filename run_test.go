package ecg_test

import (
	"embed"
	ecg "editorconfig-guesser"
	_ "editorconfig-guesser/fileformats"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/*.txtar
var testData embed.FS

func TestRunInDir(t *testing.T) {
	files, err := fs.Glob(testData, "testdata/*.txtar")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			content, err := testData.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			archive := txtar.Parse(content)

			mapFS := make(fstest.MapFS)
			var expected string

			for _, f := range archive.Files {
				if f.Name == "expected.editorconfig" {
					expected = string(f.Data)
					continue
				}
				mapFS[f.Name] = &fstest.MapFile{
					Data: f.Data,
				}
			}

			ignore := func(f *ecg.File) bool {
				return false
			}

			got, err := ecg.RunInDir(mapFS, ignore)
			if err != nil {
				t.Fatalf("RunInDir failed: %v", err)
			}

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("RunInDir() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
