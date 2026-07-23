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
			if before, ok := strings.CutSuffix(file.Name, ".expect.txt"); ok {
				name := before
				expects[name] = string(file.Data)
			} else if before, ok := strings.CutSuffix(file.Name, ".txt"); ok {
				name := before
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
				gen := html.NewGenerator(html.WithDataPositions())
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

func TestGeneratorWithDataPositions(t *testing.T) {
	root, err := a4code.ParseString(`[quoteof "User" Hi]`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	var buf bytes.Buffer
	gen := html.NewGenerator(html.WithDataPositions())
	if err := ast.Generate(&buf, root, gen); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `<blockquote class="a4code-block a4code-quoteof quote-color-0" data-start-pos="0" data-end-pos="2">`) {
		t.Fatalf("missing quote data positions in %q", got)
	}
	if !strings.Contains(got, `<span data-start-pos="0" data-end-pos="2">Hi</span>`) {
		t.Fatalf("missing text data positions in %q", got)
	}
}

func TestGeneratorWithoutSourceAttrs(t *testing.T) {
	root, err := a4code.ParseString(`[b Bold] [img=image.jpg]`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	var buf bytes.Buffer
	gen := html.NewGenerator()
	if err := ast.Generate(&buf, root, gen); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, `data-start-pos=`) || strings.Contains(got, `data-end-pos=`) {
		t.Fatalf("source attributes should be omitted from %q", got)
	}
	if got != `<strong><span>Bold</span></strong><span> </span><img src="image.jpg" />` {
		t.Fatalf("unexpected markup without source positions: %q", got)
	}
}

type testSourceAttr struct{}

func (testSourceAttr) SourceAttrs(start, end int) string {
	return ` data-test-source="yes"`
}

func TestGeneratorWithCustomSourceAttrBuilder(t *testing.T) {
	root, err := a4code.ParseString(`Hi`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	var buf bytes.Buffer
	gen := html.NewGenerator(html.WithDataPositions(), html.WithSourceAttrBuilder(testSourceAttr{}))
	if err := ast.Generate(&buf, root, gen); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := buf.String()
	if got != `<span data-start-pos="0" data-end-pos="2" data-test-source="yes">Hi</span>` {
		t.Fatalf("unexpected markup with custom source attribute: %q", got)
	}
}
