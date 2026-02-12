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
	io.WriteString(w, t.Value)
	return nil
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	io.WriteString(w, "**")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "**")
	return nil
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	io.WriteString(w, "*")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "*")
	return nil
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	io.WriteString(w, "<u>")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</u>")
	return nil
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	io.WriteString(w, "<sup>")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</sup>")
	return nil
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	io.WriteString(w, "<sub>")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</sub>")
	return nil
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	io.WriteString(w, "[")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "](%s)", n.Href)
	return nil
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	fmt.Fprintf(w, "![](%s)", n.Src)
	return nil
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	io.WriteString(w, "\n```\n")
	io.WriteString(w, n.Value)
	io.WriteString(w, "\n```\n")
	return nil
}

func (g *Generator) CodeIn(w io.Writer, n *ast.CodeIn) error {
	io.WriteString(w, "\n```")
	io.WriteString(w, n.Language)
	io.WriteString(w, "\n")
	io.WriteString(w, n.Value)
	io.WriteString(w, "\n```\n")
	return nil
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	// Fallback to HTML for complex blocks
	io.WriteString(w, "<blockquote>")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</blockquote>")
	return nil
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	fmt.Fprintf(w, "<blockquote><p>Quote of %s:</p>", n.Name)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</blockquote>")
	return nil
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	io.WriteString(w, "<details><summary>Spoiler</summary>")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</details>")
	return nil
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	io.WriteString(w, "<blockquote>")
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</blockquote>")
	return nil
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	if sw, ok := w.(*SmartWriter); ok {
		if sw.lastByte != '\n' {
			io.WriteString(w, "\n")
		}
	}
	io.WriteString(w, "---\n")
	return nil
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	// Just output children for custom tags
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return nil
}
