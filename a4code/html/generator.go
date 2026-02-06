package html

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/url"

	"github.com/arran4/goa4web/a4code/ast"
)

type Generator struct {
	Depth int
}

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
	fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, t.Start, t.End)
	for i := 0; i < len(t.Value); i++ {
		switch t.Value[i] {
		case '&':
			io.WriteString(w, "&amp;")
		case '<':
			io.WriteString(w, "&lt;")
		case '>':
			io.WriteString(w, "&gt;")
		case '\r':
		case '\n':
			io.WriteString(w, "<br />\n")
		default:
			writeByte(w, t.Value[i])
		}
	}
	io.WriteString(w, "</span>")
	return nil
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	fmt.Fprintf(w, `<strong data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</strong>")
	return nil
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	fmt.Fprintf(w, `<i data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</i>")
	return nil
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	fmt.Fprintf(w, `<u data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</u>")
	return nil
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	fmt.Fprintf(w, `<sup data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</sup>")
	return nil
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	fmt.Fprintf(w, `<sub data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</sub>")
	return nil
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	if safe, ok := SanitizeURL(n.Href); ok {
		fmt.Fprintf(w, `<a href="`)
		io.WriteString(w, safe)
		fmt.Fprintf(w, `" target="_BLANK" data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
		for _, c := range n.Children {
			if err := ast.Generate(w, c, g); err != nil {
				return err
			}
		}
		io.WriteString(w, "</a>")
	} else {
		fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
		io.WriteString(w, safe)
		for _, c := range n.Children {
			if err := ast.Generate(w, c, g); err != nil {
				return err
			}
		}
		io.WriteString(w, "</span>")
	}
	return nil
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	io.WriteString(w, "<img src=\"")
	io.WriteString(w, htmlEscape(n.Src))
	fmt.Fprintf(w, `" data-start-pos="%d" data-end-pos="%d" />`, n.Start, n.End)
	return nil
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	fmt.Fprintf(w, `<pre class="a4code-block a4code-code" data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, n.InnerStart, n.InnerEnd)
	io.WriteString(w, htmlEscape(n.Value))
	io.WriteString(w, "</span></pre>")
	return nil
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	colorClass := fmt.Sprintf("quote-color-%d", g.Depth%6)
	fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quote %s" data-start-pos="%d" data-end-pos="%d">`, colorClass, n.Start, n.End)
	io.WriteString(w, "<div class=\"quote-body\">")

	childGen := &Generator{Depth: g.Depth + 1}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, childGen); err != nil {
			return err
		}
	}
	io.WriteString(w, "</div>")
	io.WriteString(w, "</blockquote>")
	return nil
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	colorClass := fmt.Sprintf("quote-color-%d", g.Depth%6)
	fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quoteof %s" data-start-pos="%d" data-end-pos="%d">`, colorClass, n.Start, n.End)
	io.WriteString(w, "<div class=\"quote-header\">Quote of ")
	io.WriteString(w, htmlEscape(n.Name))
	io.WriteString(w, ":</div>")
	io.WriteString(w, "<div class=\"quote-body\">")

	childGen := &Generator{Depth: g.Depth + 1}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, childGen); err != nil {
			return err
		}
	}
	io.WriteString(w, "</div>")
	io.WriteString(w, "</blockquote>")
	return nil
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	fmt.Fprintf(w, `<span class="spoiler" data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</span>")
	return nil
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	fmt.Fprintf(w, `<div class="a4code-block a4code-indent" data-start-pos="%d" data-end-pos="%d"><div>`, n.Start, n.End)
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "</div></div>")
	return nil
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	fmt.Fprintf(w, `<hr data-start-pos="%d" data-end-pos="%d" />`, n.Start, n.End)
	return nil
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, n.Start, n.End)
	io.WriteString(w, "[")
	io.WriteString(w, htmlEscape(n.Tag))
	for _, ch := range n.Children {
		if err := ast.Generate(w, ch, g); err != nil {
			return err
		}
	}
	io.WriteString(w, "]")
	io.WriteString(w, "</span>")
	return nil
}

func writeByte(w io.Writer, b byte) {
	if bw, ok := w.(io.ByteWriter); ok {
		_ = bw.WriteByte(b)
		return
	}
	_, _ = w.Write([]byte{b})
}

func htmlEscape(s string) string {
	var b bytes.Buffer
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func SanitizeURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return html.EscapeString(raw), false
	}
	switch u.Scheme {
	case "http", "https":
		return html.EscapeString(u.String()), true
	default:
		return html.EscapeString(raw), false
	}
}
