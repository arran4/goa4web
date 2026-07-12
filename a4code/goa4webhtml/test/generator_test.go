package goa4webhtml_test

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/a4code/goa4webhtml"
	"golang.org/x/tools/txtar"
)

//go:embed tests.txtar
var testData embed.FS

type mockLinkProvider struct{}

func (m *mockLinkProvider) RenderLink(url string, isBlock bool, isImmediateClose bool) (htmlOpen string, htmlClose string, consumeImmediate bool) {
	return fmt.Sprintf(`<custom-link href="%s">`, url), "</custom-link>", false
}

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
				// Trim trailing newline from input to avoid implicit Text node generation
				input := strings.TrimSuffix(inputRaw, "\n")

				expect, ok := expects[name]
				if !ok {
					t.Fatalf("No expectation found for %s", name)
				}
				// Expectation might also have a trailing newline from txtar, trim it to match exact output
				expect = strings.TrimSuffix(expect, "\n")

				root, err := a4code.ParseString(input)
				if err != nil {
					t.Fatalf("ParseString error: %v", err)
				}

				var buf bytes.Buffer
				// Use mocked providers for specific tests via With... options
				var opts []interface{}
				if strings.Contains(name, "provider_link") {
					opts = append(opts, goa4webhtml.WithLinkProvider(&mockLinkProvider{}))
				}
				if strings.Contains(name, "provider_image") {
					opts = append(opts, goa4webhtml.WithImageMapper(func(tag, val string) string {
						return "/mapped/" + val
					}))
				}
				if strings.Contains(name, "provider_quoteof") {
					opts = append(opts, goa4webhtml.WithUserColorMapper(func(username string) string {
						return "user-color-" + username
					}))
				}

				gen := goa4webhtml.NewGenerator(opts...)
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

func TestGeneratorWithDataOffset(t *testing.T) {
	root, err := a4code.ParseString(`[quoteof "User" [img=image.jpg]]`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	var buf bytes.Buffer
	gen := goa4webhtml.NewGenerator(
		goa4webhtml.WithImageMapper(func(tag, val string) string {
			return "/mapped/" + val
		}),
		goa4webhtml.WithUserColorMapper(func(username string) string {
			return "user-color-" + username
		}),
		goa4webhtml.WithDataOffset(),
	)
	if err := ast.Generate(&buf, root, gen); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `<blockquote class="a4code-block a4code-quoteof user-color-User quote-color-0" data-offset="0" data-start-pos="0" data-end-pos="1">`) {
		t.Fatalf("missing quote data-offset in %q", got)
	}
	if !strings.Contains(got, `<img src="/mapped/image.jpg" data-offset="0" data-start-pos="0" data-end-pos="1" />`) {
		t.Fatalf("missing mapped image data-offset in %q", got)
	}
}

func TestGeneratorWithoutDataPositions(t *testing.T) {
	root, err := a4code.ParseString(`[quoteof "User" [img=image.jpg]]`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	var buf bytes.Buffer
	gen := goa4webhtml.NewGenerator(
		goa4webhtml.WithImageMapper(func(tag, val string) string {
			return "/mapped/" + val
		}),
		goa4webhtml.WithoutDataPositions(),
	)
	if err := ast.Generate(&buf, root, gen); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, `data-start-pos=`) || strings.Contains(got, `data-end-pos=`) {
		t.Fatalf("source position attributes should be omitted from %q", got)
	}
	if !strings.Contains(got, `<blockquote class="a4code-block a4code-quoteof quote-color-0">`) {
		t.Fatalf("quote markup should remain without source positions in %q", got)
	}
	if !strings.Contains(got, `<img src="/mapped/image.jpg" />`) {
		t.Fatalf("mapped image markup should remain without source positions in %q", got)
	}
}
