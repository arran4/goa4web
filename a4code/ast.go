package a4code

import (
	"bytes"
	"io"
)

func writeByte(w io.Writer, b byte) {
	w.Write([]byte{b})
}

// Node represents a parsed element of markup.
type Node interface {
	html(io.Writer)
	a4code(io.Writer)
	isNode()
}

type parent interface {
	childrenPtr() *[]Node
}

// Root is the top level node of a document.
type Root struct {
	Children []Node
}

func (*Root) isNode() {}

func (r *Root) html(w io.Writer) {
	for _, c := range r.Children {
		c.html(w)
	}
}

func (r *Root) a4code(w io.Writer) {
	for _, c := range r.Children {
		c.a4code(w)
	}
}

func (r *Root) childrenPtr() *[]Node { return &r.Children }

// Text contains plain text content.
type Text struct {
	Value string
}

func (*Text) isNode() {}

func (t *Text) html(w io.Writer) {
	for i := 0; i < len(t.Value); i++ {
		switch t.Value[i] {
		case '&':
			io.WriteString(w, "&amp;")
		case '<':
			io.WriteString(w, "&lt;")
		case '>':
			io.WriteString(w, "&gt;")
		case '\n':
			io.WriteString(w, "<br />\n")
		default:
			writeByte(w, t.Value[i])
		}
	}
}

func (t *Text) a4code(w io.Writer) {
	for i := 0; i < len(t.Value); i++ {
		switch t.Value[i] {
		case '[', ']', '=', '\\', '*', '/', '_':
			writeByte(w, '\\')
			writeByte(w, t.Value[i])
		default:
			writeByte(w, t.Value[i])
		}
	}
}

// Bold text.
type Bold struct{ Children []Node }

func (*Bold) isNode()                {}
func (b *Bold) childrenPtr() *[]Node { return &b.Children }

func (b *Bold) html(w io.Writer) {
	io.WriteString(w, "<strong>")
	for _, c := range b.Children {
		c.html(w)
	}
	io.WriteString(w, "</strong>")
}

func (b *Bold) a4code(w io.Writer) {
	io.WriteString(w, "[b")
	for _, c := range b.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Italic text.
type Italic struct{ Children []Node }

func (*Italic) isNode()                {}
func (i *Italic) childrenPtr() *[]Node { return &i.Children }

func (i *Italic) html(w io.Writer) {
	io.WriteString(w, "<i>")
	for _, c := range i.Children {
		c.html(w)
	}
	io.WriteString(w, "</i>")
}

func (i *Italic) a4code(w io.Writer) {
	io.WriteString(w, "[i")
	for _, c := range i.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Underline text.
type Underline struct{ Children []Node }

func (*Underline) isNode()                {}
func (u *Underline) childrenPtr() *[]Node { return &u.Children }

func (u *Underline) html(w io.Writer) {
	io.WriteString(w, "<u>")
	for _, c := range u.Children {
		c.html(w)
	}
	io.WriteString(w, "</u>")
}

func (u *Underline) a4code(w io.Writer) {
	io.WriteString(w, "[u")
	for _, c := range u.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Superscript text.
type Sup struct{ Children []Node }

func (*Sup) isNode()                {}
func (s *Sup) childrenPtr() *[]Node { return &s.Children }

func (s *Sup) html(w io.Writer) {
	io.WriteString(w, "<sup>")
	for _, c := range s.Children {
		c.html(w)
	}
	io.WriteString(w, "</sup>")
}

func (s *Sup) a4code(w io.Writer) {
	io.WriteString(w, "[sup")
	for _, c := range s.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Subscript text.
type Sub struct{ Children []Node }

func (*Sub) isNode()                {}
func (s *Sub) childrenPtr() *[]Node { return &s.Children }

func (s *Sub) html(w io.Writer) {
	io.WriteString(w, "<sub>")
	for _, c := range s.Children {
		c.html(w)
	}
	io.WriteString(w, "</sub>")
}

func (s *Sub) a4code(w io.Writer) {
	io.WriteString(w, "[sub")
	for _, c := range s.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Link to a URL.
type Link struct {
	Href     string
	Children []Node
}

func (*Link) isNode()                {}
func (l *Link) childrenPtr() *[]Node { return &l.Children }

func (l *Link) html(w io.Writer) {
	if safe, ok := SanitizeURL(l.Href); ok {
		io.WriteString(w, "<a href=\"")
		io.WriteString(w, safe)
		io.WriteString(w, "\" target=\"_BLANK\">")
		for _, c := range l.Children {
			c.html(w)
		}
		io.WriteString(w, "</a>")
	} else {
		io.WriteString(w, safe)
		for _, c := range l.Children {
			c.html(w)
		}
	}
}

func (l *Link) a4code(w io.Writer) {
	io.WriteString(w, "[a=")
	escapeArg(w, l.Href)
	for _, c := range l.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Image embeds an image.
type Image struct{ Src string }

func (*Image) isNode() {}

func (i *Image) html(w io.Writer) {
	io.WriteString(w, "<img src=\"")
	io.WriteString(w, htmlEscape(i.Src))
	io.WriteString(w, "\" />")
}

func (i *Image) a4code(w io.Writer) {
	io.WriteString(w, "[img=")
	escapeArg(w, i.Src)
	writeByte(w, ']')
}

// Code block.
type Code struct{ Value string }

func (*Code) isNode() {}

func (c *Code) html(w io.Writer) {
	io.WriteString(w, "<pre class=\"a4code-block a4code-code\">")
	io.WriteString(w, htmlEscape(c.Value))
	io.WriteString(w, "</pre>")
}

func (c *Code) a4code(w io.Writer) {
	io.WriteString(w, "[code]")
	io.WriteString(w, c.Value)
	io.WriteString(w, "[/code]")
}

// Quote node.
type Quote struct{ Children []Node }

func (*Quote) isNode()                {}
func (q *Quote) childrenPtr() *[]Node { return &q.Children }

func (q *Quote) html(w io.Writer) {
	io.WriteString(w, "<blockquote class=\"a4code-block a4code-quote\">")
	for _, c := range q.Children {
		c.html(w)
	}
	io.WriteString(w, "</blockquote>")
}

func (q *Quote) a4code(w io.Writer) {
	io.WriteString(w, "[quote")
	for _, c := range q.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// QuoteOf node.
type QuoteOf struct {
	Name     string
	Children []Node
}

func (*QuoteOf) isNode()                {}
func (q *QuoteOf) childrenPtr() *[]Node { return &q.Children }

func (q *QuoteOf) html(w io.Writer) {
	io.WriteString(w, "<blockquote class=\"a4code-block a4code-quoteof\"><div>Quote of ")
	io.WriteString(w, htmlEscape(q.Name))
	io.WriteString(w, ":</div>")
	for _, c := range q.Children {
		c.html(w)
	}
	io.WriteString(w, "</blockquote>")
}

func (q *QuoteOf) a4code(w io.Writer) {
	io.WriteString(w, "[quoteof ")
	escapeArg(w, q.Name)
	for _, c := range q.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Spoiler node.
type Spoiler struct{ Children []Node }

func (*Spoiler) isNode()                {}
func (s *Spoiler) childrenPtr() *[]Node { return &s.Children }

func (s *Spoiler) html(w io.Writer) {
	io.WriteString(w, "<span class=\"spoiler\">")
	for _, c := range s.Children {
		c.html(w)
	}
	io.WriteString(w, "</span>")
}

func (s *Spoiler) a4code(w io.Writer) {
	io.WriteString(w, "[spoiler")
	for _, c := range s.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// Indent node.
type Indent struct{ Children []Node }

func (*Indent) isNode()                {}
func (i *Indent) childrenPtr() *[]Node { return &i.Children }

func (i *Indent) html(w io.Writer) {
	io.WriteString(w, "<div class=\"a4code-block a4code-indent\"><div>")
	for _, c := range i.Children {
		c.html(w)
	}
	io.WriteString(w, "</div></div>")
}

func (i *Indent) a4code(w io.Writer) {
	io.WriteString(w, "[indent")
	for _, c := range i.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

// HR node.
type HR struct{}

func (*HR) isNode() {}

func (*HR) html(w io.Writer) { io.WriteString(w, "<hr/>") }

func (*HR) a4code(w io.Writer) { io.WriteString(w, "[hr]") }

// Custom element for unrecognised tags.
type Custom struct {
	Tag      string
	Children []Node
}

func (*Custom) isNode()                {}
func (c *Custom) childrenPtr() *[]Node { return &c.Children }

func (c *Custom) html(w io.Writer) {
	io.WriteString(w, "[")
	io.WriteString(w, htmlEscape(c.Tag))
	for _, ch := range c.Children {
		ch.html(w)
	}
	io.WriteString(w, "]")
}

func (c *Custom) a4code(w io.Writer) {
	writeByte(w, '[')
	io.WriteString(w, c.Tag)
	for _, ch := range c.Children {
		ch.a4code(w)
	}
	writeByte(w, ']')
}

// helper to escape plain strings for HTML output.
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

// escapeArg escapes characters in a tag argument.
func escapeArg(w io.Writer, s string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[', ']', '=', '\\':
			writeByte(w, '\\')
			writeByte(w, s[i])
		default:
			writeByte(w, s[i])
		}
	}
}
