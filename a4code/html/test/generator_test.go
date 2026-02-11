package html_test

import (
	"bytes"
	"embed"
	"io/fs"
	"strings"
	"testing"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/a4code/html"
	"golang.org/x/tools/txtar"
)

//go:embed tests.txtar
var testData embed.FS

func TestGenerator(t *testing.T) {
	fs.WalkDir(testData, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".txtar") {
			return nil
		}

		data, err := testData.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", path, err)
		}

		archive := txtar.Parse(data)
		inputs := make(map[string]string)
		expects := make(map[string]string)

		for _, file := range archive.Files {
			if strings.HasSuffix(file.Name, ".expect.txt") {
				name := strings.TrimSuffix(file.Name, ".expect.txt")
				expects[name] = string(file.Data)
			} else if strings.HasSuffix(file.Name, ".txt") {
				name := strings.TrimSuffix(file.Name, ".txt")
				inputs[name] = string(file.Data)
			}
		}

		for name, inputRaw := range inputs {
			t.Run(name, func(t *testing.T) {
				input := strings.TrimSuffix(inputRaw, "\n")
				expect, ok := expects[name]
				if !ok {
					t.Fatalf("No expectation found for %s", name)
				}
				expect = strings.TrimSuffix(expect, "\n")

				root, err := a4code.ParseString(input)
				if err != nil {
					t.Fatalf("ParseString error: %v", err)
				}

				var buf bytes.Buffer
				gen := html.NewGenerator()
				if err := ast.Generate(&buf, root, gen); err != nil {
					t.Fatalf("Generate error: %v", err)
				}

				got := buf.String()
				if got != expect {
					t.Errorf("Mismatch for %s:\nInput: %q\nGot:    %q\nWant:   %q", name, input, got, expect)
				}
			})
		}
		return nil
	})
}
