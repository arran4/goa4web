package ast

import "io"

// Generator defines how to generate output for each node type.
// Implementations should write to w.
type Generator interface {
	Root(w io.Writer, n *Root) error
	Text(w io.Writer, n *Text) error
	Bold(w io.Writer, n *Bold) error
	Italic(w io.Writer, n *Italic) error
	Underline(w io.Writer, n *Underline) error
	Sup(w io.Writer, n *Sup) error
	Sub(w io.Writer, n *Sub) error
	Link(w io.Writer, n *Link) error
	Image(w io.Writer, n *Image) error
	Code(w io.Writer, n *Code) error
	CodeIn(w io.Writer, n *CodeIn) error
	Quote(w io.Writer, n *Quote) error
	QuoteOf(w io.Writer, n *QuoteOf) error
	Spoiler(w io.Writer, n *Spoiler) error
	Indent(w io.Writer, n *Indent) error
	HR(w io.Writer, n *HR) error
	Custom(w io.Writer, n *Custom) error
}

// Generate traverses the AST and calls the appropriate Generator method.
func Generate(w io.Writer, n Node, g Generator) error {
	if n == nil {
		return nil
	}
	switch t := n.(type) {
	case *Root:
		return g.Root(w, t)
	case *Text:
		return g.Text(w, t)
	case *Bold:
		return g.Bold(w, t)
	case *Italic:
		return g.Italic(w, t)
	case *Underline:
		return g.Underline(w, t)
	case *Sup:
		return g.Sup(w, t)
	case *Sub:
		return g.Sub(w, t)
	case *Link:
		return g.Link(w, t)
	case *Image:
		return g.Image(w, t)
	case *Code:
		return g.Code(w, t)
	case *CodeIn:
		return g.CodeIn(w, t)
	case *Quote:
		return g.Quote(w, t)
	case *QuoteOf:
		return g.QuoteOf(w, t)
	case *Spoiler:
		return g.Spoiler(w, t)
	case *Indent:
		return g.Indent(w, t)
	case *HR:
		return g.HR(w, t)
	case *Custom:
		return g.Custom(w, t)
	}
	return nil
}
