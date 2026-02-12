package format

import (
	"io"
	"strings"

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
	for i := 0; i < len(t.Value); i++ {
		switch t.Value[i] {
		case '[', ']', '=', '\\', '*', '/', '_':
			writeByte(w, '\\')
			writeByte(w, t.Value[i])
		default:
			writeByte(w, t.Value[i])
		}
	}
	return nil
}

func (g *Generator) generateChildren(w io.Writer, children []ast.Node) error {
	for _, c := range children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	writeByte(w, ']')
	return nil
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	io.WriteString(w, "[b")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	io.WriteString(w, "[i")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	io.WriteString(w, "[u")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	io.WriteString(w, "[sup")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	io.WriteString(w, "[sub")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	io.WriteString(w, "[a=")
	escapeArg(w, n.Href)
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	io.WriteString(w, "[img=")
	escapeArg(w, n.Src)
	writeByte(w, ']')
	return nil
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	io.WriteString(w, "[code")
	if n.IsBlock {
		io.WriteString(w, "\n")
	} else if len(n.Value) > 0 {
		first := n.Value[0]
		if first != ' ' && first != '\n' && first != '\r' {
			io.WriteString(w, " ")
		}
	}
	io.WriteString(w, n.Value)
	io.WriteString(w, "]")
	return nil
}

func (g *Generator) CodeIn(w io.Writer, n *ast.CodeIn) error {
	io.WriteString(w, "[codein ")
	escapeQuotedArg(w, n.Language)
	if strings.Contains(n.Value, "\n") {
		writeByte(w, '\n')
	} else {
		writeByte(w, ' ')
	}
	io.WriteString(w, n.Value)
	writeByte(w, ']')
	return nil
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	io.WriteString(w, "[quote")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	io.WriteString(w, "[quoteof ")
	escapeQuotedArg(w, n.Name)
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	io.WriteString(w, "[spoiler")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	io.WriteString(w, "[indent")
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	io.WriteString(w, "[hr]")
	return nil
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	writeByte(w, '[')
	io.WriteString(w, n.Tag)
	if len(n.Children) > 0 {
		writeByte(w, ' ')
	}
	return g.generateChildren(w, n.Children)
}

func writeByte(w io.Writer, b byte) {
	if bw, ok := w.(io.ByteWriter); ok {
		_ = bw.WriteByte(b)
		return
	}
	_, _ = w.Write([]byte{b})
}

func escapeArg(w io.Writer, s string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[', ']', '=', '\\', ' ': // Escape space too now!
			writeByte(w, '\\')
			writeByte(w, s[i])
		default:
			writeByte(w, s[i])
		}
	}
}

func escapeQuotedArg(w io.Writer, s string) {
	writeByte(w, '"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"', '\\':
			writeByte(w, '\\')
			writeByte(w, s[i])
		default:
			writeByte(w, s[i])
		}
	}
	writeByte(w, '"')
}
