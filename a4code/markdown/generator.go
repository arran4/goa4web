package markdown

import (
	"fmt"
	"io"

	"github.com/arran4/goa4web/a4code/ast"
)

type SmartWriter struct {
	w        io.Writer
	lastByte byte
}

func (sw *SmartWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	n, err = sw.w.Write(p)
	if n > 0 {
		sw.lastByte = p[n-1]
	}
	return
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func writeString(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}

func (g *Generator) Root(w io.Writer, n *ast.Root) error {
	sw := &SmartWriter{w: w, lastByte: '\n'}
	for _, c := range n.Children {
		if err := ast.Generate(sw, c, g); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) Text(w io.Writer, t *ast.Text) error {
	return writeString(w, t.Value)
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	if err := writeString(w, "**"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "**")
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	if err := writeString(w, "*"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "*")
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	if err := writeString(w, "<u>"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</u>")
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	if err := writeString(w, "<sup>"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</sup>")
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	if err := writeString(w, "<sub>"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</sub>")
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	if err := writeString(w, "["); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "](%s)", n.Href)
	return err
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	_, err := fmt.Fprintf(w, "![](%s)", n.Src)
	return err
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	if err := writeString(w, "\n```\n"); err != nil {
		return err
	}
	if err := writeString(w, n.Value); err != nil {
		return err
	}
	return writeString(w, "\n```\n")
}

func (g *Generator) CodeIn(w io.Writer, n *ast.CodeIn) error {
	if err := writeString(w, "\n```"); err != nil {
		return err
	}
	if err := writeString(w, n.Language); err != nil {
		return err
	}
	if err := writeString(w, "\n"); err != nil {
		return err
	}
	if err := writeString(w, n.Value); err != nil {
		return err
	}
	return writeString(w, "\n```")
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	if err := writeString(w, "<blockquote>"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</blockquote>")
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	if _, err := fmt.Fprintf(w, "<blockquote><p>Quote of %s:</p>", n.Name); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</blockquote>")
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	if err := writeString(w, "<details><summary>Spoiler</summary>"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</details>")
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	if err := writeString(w, "<blockquote>"); err != nil {
		return err
	}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return writeString(w, "</blockquote>")
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	if sw, ok := w.(*SmartWriter); ok {
		if sw.lastByte != '\n' {
			if err := writeString(w, "\n"); err != nil {
				return err
			}
		}
	}
	return writeString(w, "---\n")
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return nil
}
