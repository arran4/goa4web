package html

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/a4code/ast"
)

type Generator struct {
	Depth              int
	Self               ast.Generator
	SourceAttrBuilders []SourceAttrBuilder
}

// SourceAttrBuilder renders source-related HTML attributes for an AST node span.
type SourceAttrBuilder interface {
	SourceAttrs(start, end int) string
}

// Option configures an HTML generator.
type Option func(*Generator)

// WithDataPositions emits data-start-pos and data-end-pos attributes.
func WithDataPositions() Option {
	return WithSourceAttrBuilder(DataPositionAttrs{})
}

// WithSourceAttrBuilder appends a source-related attribute builder.
func WithSourceAttrBuilder(builder SourceAttrBuilder) Option {
	return func(g *Generator) { g.SourceAttrBuilders = append(g.sourceAttrBuilders(), builder) }
}

func NewGenerator(opts ...Option) *Generator {
	g := &Generator{}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Generator) self() ast.Generator {
	if g.Self != nil {
		return g.Self
	}
	return g
}

func (g *Generator) Root(w io.Writer, n *ast.Root) error {
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) Text(w io.Writer, t *ast.Text) error {
	_, _ = fmt.Fprintf(w, `<span%s>`, g.SourceAttrs(t.Start, t.End))
	for i := 0; i < len(t.Value); i++ {
		switch t.Value[i] {
		case '&':
			_, _ = io.WriteString(w, "&amp;")
		case '<':
			_, _ = io.WriteString(w, "&lt;")
		case '>':
			_, _ = io.WriteString(w, "&gt;")
		case '\r':
		case '\n':
			_, _ = io.WriteString(w, "<br />\n")
		default:
			writeByte(w, t.Value[i])
		}
	}
	_, _ = io.WriteString(w, "</span>")
	return nil
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	_, _ = fmt.Fprintf(w, `<strong%s>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</strong>")
	return nil
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	_, _ = fmt.Fprintf(w, `<i%s>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</i>")
	return nil
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	_, _ = fmt.Fprintf(w, `<u%s>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</u>")
	return nil
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	_, _ = fmt.Fprintf(w, `<sup%s>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</sup>")
	return nil
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	_, _ = fmt.Fprintf(w, `<sub%s>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</sub>")
	return nil
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	if safe, ok := SanitizeURL(n.Href); ok {
		_, _ = fmt.Fprintf(w, `<a href="`)
		_, _ = io.WriteString(w, safe)
		_, _ = fmt.Fprintf(w, `" target="_BLANK"%s>`, g.SourceAttrs(n.Start, n.End))
		if isEffectivelyEmpty(n.Children) {
			_, _ = io.WriteString(w, safe)
		} else {
			for _, c := range n.Children {
				if err := ast.Generate(w, c, g.self()); err != nil {
					return err
				}
			}
		}
		_, _ = io.WriteString(w, "</a>")
	} else {
		_, _ = fmt.Fprintf(w, `<span%s>`, g.SourceAttrs(n.Start, n.End))
		_, _ = io.WriteString(w, safe)
		for _, c := range n.Children {
			if err := ast.Generate(w, c, g.self()); err != nil {
				return err
			}
		}
		_, _ = io.WriteString(w, "</span>")
	}
	return nil
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	_, _ = io.WriteString(w, "<img src=\"")
	_, _ = io.WriteString(w, htmlEscape(n.Src))
	_, _ = fmt.Fprintf(w, `"%s />`, g.SourceAttrs(n.Start, n.End))
	return nil
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	if !ast.IsBlockNode(n) {
		_, _ = fmt.Fprintf(w, `<code class="a4code-inline a4code-code"%s>`, g.SourceAttrs(n.Start, n.End))
		_, _ = io.WriteString(w, htmlEscape(n.Value))
		_, _ = io.WriteString(w, "</code>")
		return nil
	}
	_, _ = fmt.Fprintf(w, `<pre class="a4code-block a4code-code"%s>`, g.SourceAttrs(n.Start, n.End))
	_, _ = fmt.Fprintf(w, `<span%s>`, g.SourceAttrs(n.InnerStart, n.InnerEnd))
	_, _ = io.WriteString(w, htmlEscape(n.Value))
	_, _ = io.WriteString(w, "</span></pre>")
	return nil
}

func (g *Generator) CodeIn(w io.Writer, n *ast.CodeIn) error {
	if !ast.IsBlockNode(n) {
		_, _ = fmt.Fprintf(w, `<code class="a4code-inline a4code-code a4code-language-%s language-%s"%s>`, htmlEscape(n.Language), htmlEscape(n.Language), g.SourceAttrs(n.Start, n.End))
		_, _ = fmt.Fprintf(w, `<span%s>`, g.SourceAttrs(n.InnerStart, n.InnerEnd))
		_, _ = io.WriteString(w, htmlEscape(n.Value))
		_, _ = io.WriteString(w, "</span></code>")
		return nil
	}
	_, _ = fmt.Fprintf(w, `<pre class="a4code-block a4code-code a4code-language-%s"%s>`, htmlEscape(n.Language), g.SourceAttrs(n.Start, n.End))
	_, _ = fmt.Fprintf(w, `<code class="language-%s">`, htmlEscape(n.Language))
	_, _ = fmt.Fprintf(w, `<span%s>`, g.SourceAttrs(n.InnerStart, n.InnerEnd))
	_, _ = io.WriteString(w, htmlEscape(n.Value))
	_, _ = io.WriteString(w, "</span></code></pre>")
	return nil
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	if ast.IsBlockNode(n) {
		colorClass := fmt.Sprintf("quote-color-%d", g.Depth%6)
		_, _ = fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quote %s"%s>`, colorClass, g.SourceAttrs(n.Start, n.End))
		_, _ = io.WriteString(w, "<div class=\"quote-body\">")

		childGen := &Generator{Depth: g.Depth + 1, Self: g.Self, SourceAttrBuilders: g.sourceAttrBuilders()}
		for _, c := range n.Children {
			if err := ast.Generate(w, c, childGen); err != nil {
				return err
			}
		}
		_, _ = io.WriteString(w, "</div>")
		_, _ = io.WriteString(w, "</blockquote>")
	} else {
		_, _ = fmt.Fprintf(w, `<q class="a4code-inline a4code-quote"%s>`, g.SourceAttrs(n.Start, n.End))
		childGen := &Generator{Depth: g.Depth + 1, Self: g.Self, SourceAttrBuilders: g.sourceAttrBuilders()}
		for _, c := range n.Children {
			if err := ast.Generate(w, c, childGen); err != nil {
				return err
			}
		}
		_, _ = io.WriteString(w, "</q>")
	}
	return nil
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	colorClass := fmt.Sprintf("quote-color-%d", g.Depth%6)
	_, _ = fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quoteof %s"%s>`, colorClass, g.SourceAttrs(n.Start, n.End))
	_, _ = io.WriteString(w, "<div class=\"quote-header\">Quote of ")
	_, _ = io.WriteString(w, htmlEscape(n.Name))
	_, _ = io.WriteString(w, ":</div>")
	_, _ = io.WriteString(w, "<div class=\"quote-body\">")

	childGen := &Generator{Depth: g.Depth + 1, Self: g.Self, SourceAttrBuilders: g.sourceAttrBuilders()}
	for _, c := range n.Children {
		if err := ast.Generate(w, c, childGen); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</div>")
	_, _ = io.WriteString(w, "</blockquote>")
	return nil
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	_, _ = fmt.Fprintf(w, `<span class="spoiler"%s>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</span>")
	return nil
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	_, _ = fmt.Fprintf(w, `<div class="a4code-block a4code-indent"%s><div>`, g.SourceAttrs(n.Start, n.End))
	for _, c := range n.Children {
		if err := ast.Generate(w, c, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "</div></div>")
	return nil
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	_, _ = fmt.Fprintf(w, `<hr%s />`, g.SourceAttrs(n.Start, n.End))
	return nil
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	_, _ = fmt.Fprintf(w, `<span%s>`, g.SourceAttrs(n.Start, n.End))
	_, _ = io.WriteString(w, "[")
	_, _ = io.WriteString(w, htmlEscape(n.Tag))
	for _, ch := range n.Children {
		if err := ast.Generate(w, ch, g.self()); err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "]")
	_, _ = io.WriteString(w, "</span>")
	return nil
}

// SourceAttrs returns enabled source-related attributes for start and end offsets.
func (g *Generator) SourceAttrs(start, end int) string {
	var attrs strings.Builder
	for _, builder := range g.sourceAttrBuilders() {
		attrs.WriteString(builder.SourceAttrs(start, end))
	}
	return attrs.String()
}

func (g *Generator) sourceAttrBuilders() []SourceAttrBuilder {
	return g.SourceAttrBuilders
}

// DataPositionAttrs renders data-start-pos and data-end-pos attributes.
type DataPositionAttrs struct{}

// SourceAttrs renders data-start-pos and data-end-pos attributes.
func (DataPositionAttrs) SourceAttrs(start, end int) string {
	return fmt.Sprintf(` data-start-pos="%d" data-end-pos="%d"`, start, end)
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

func isEffectivelyEmpty(children []ast.Node) bool {
	if len(children) == 0 {
		return true
	}
	for _, c := range children {
		if t, ok := c.(*ast.Text); ok {
			if strings.TrimSpace(t.Value) != "" {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
