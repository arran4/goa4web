package text

import (
	"io"

	"github.com/arran4/goa4web/a4code/ast"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Root(w io.Writer, n *ast.Root) error {
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) Text(w io.Writer, t *ast.Text) error {
	io.WriteString(w, t.Value)
	return nil
}

func (g *Generator) visitChildren(w io.Writer, children []ast.Node) error {
	for _, c := range children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	return nil
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	return nil
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	return nil
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	return g.visitChildren(w, n.Children)
}
