package a4code

import (
	"bytes"
	"fmt"
	"io"
)

func writeByte(w io.Writer, b byte) {
	w.Write([]byte{b})
}

// Node represents a parsed element of markup.
type Node interface {
	html(io.Writer, int)
	a4code(io.Writer)
	isNode()
	Transform(op func(Node) (Node, error)) (Node, error)
	SetPos(start, end int)
	GetPos() (int, int)
}

type BaseNode struct {
	Start int
	End   int
}

func (n *BaseNode) SetPos(start, end int) {
	n.Start = start
	n.End = end
}

func (n *BaseNode) GetPos() (int, int) {
	return n.Start, n.End
}

type parent interface {
	childrenPtr() *[]Node
}

// Walk traverses the node tree depth-first without modifying nodes.
func Walk(n Node, fn func(Node) error) error {
	if n == nil {
		return nil
	}
	if err := fn(n); err != nil {
		return err
	}
	if p, ok := n.(parent); ok {
		for _, c := range *p.childrenPtr() {
			if err := Walk(c, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Root is the top level node of a document.
type Root struct {
	BaseNode
	Children []Node
}

func (*Root) isNode() {}

func (r *Root) html(w io.Writer, depth int) {
	for _, c := range r.Children {
		c.html(w, depth)
	}
}

func (r *Root) a4code(w io.Writer) {
	for _, c := range r.Children {
		c.a4code(w)
	}
}

func (r *Root) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := r.Children[:0]
	for _, c := range r.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	r.Children = newChildren
	return op(r)
}

func (r *Root) childrenPtr() *[]Node { return &r.Children }

// Text contains plain text content.
type Text struct {
	BaseNode
	Value string
}

func (*Text) isNode() {}

func (t *Text) html(w io.Writer, depth int) {
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

func (t *Text) Transform(op func(Node) (Node, error)) (Node, error) {
	return op(t)
}

// Bold text.
type Bold struct {
	BaseNode
	Children []Node
}

func (*Bold) isNode()                {}
func (b *Bold) childrenPtr() *[]Node { return &b.Children }

func (b *Bold) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<strong data-start-pos="%d" data-end-pos="%d">`, b.Start, b.End)
	for _, c := range b.Children {
		c.html(w, depth)
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

func (b *Bold) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := b.Children[:0]
	for _, c := range b.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	b.Children = newChildren
	return op(b)
}

// Italic text.
type Italic struct {
	BaseNode
	Children []Node
}

func (*Italic) isNode()                {}
func (i *Italic) childrenPtr() *[]Node { return &i.Children }

func (i *Italic) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<i data-start-pos="%d" data-end-pos="%d">`, i.Start, i.End)
	for _, c := range i.Children {
		c.html(w, depth)
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

func (i *Italic) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := i.Children[:0]
	for _, c := range i.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	i.Children = newChildren
	return op(i)
}

// Underline text.
type Underline struct {
	BaseNode
	Children []Node
}

func (*Underline) isNode()                {}
func (u *Underline) childrenPtr() *[]Node { return &u.Children }

func (u *Underline) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<u data-start-pos="%d" data-end-pos="%d">`, u.Start, u.End)
	for _, c := range u.Children {
		c.html(w, depth)
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

func (u *Underline) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := u.Children[:0]
	for _, c := range u.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	u.Children = newChildren
	return op(u)
}

// Superscript text.
type Sup struct {
	BaseNode
	Children []Node
}

func (*Sup) isNode()                {}
func (s *Sup) childrenPtr() *[]Node { return &s.Children }

func (s *Sup) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<sup data-start-pos="%d" data-end-pos="%d">`, s.Start, s.End)
	for _, c := range s.Children {
		c.html(w, depth)
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

func (s *Sup) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := s.Children[:0]
	for _, c := range s.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	s.Children = newChildren
	return op(s)
}

// Subscript text.
type Sub struct {
	BaseNode
	Children []Node
}

func (*Sub) isNode()                {}
func (s *Sub) childrenPtr() *[]Node { return &s.Children }

func (s *Sub) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<sub data-start-pos="%d" data-end-pos="%d">`, s.Start, s.End)
	for _, c := range s.Children {
		c.html(w, depth)
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

func (s *Sub) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := s.Children[:0]
	for _, c := range s.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	s.Children = newChildren
	return op(s)
}

// Link to a URL.
type Link struct {
	BaseNode
	Href     string
	Children []Node
}

func (*Link) isNode()                {}
func (l *Link) childrenPtr() *[]Node { return &l.Children }

func (l *Link) html(w io.Writer, depth int) {
	if safe, ok := SanitizeURL(l.Href); ok {
		fmt.Fprintf(w, `<a href="`)
		io.WriteString(w, safe)
		fmt.Fprintf(w, `" target="_BLANK" data-start-pos="%d" data-end-pos="%d">`, l.Start, l.End)
		for _, c := range l.Children {
			c.html(w, depth)
		}
		io.WriteString(w, "</a>")
	} else {
		fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, l.Start, l.End)
		io.WriteString(w, safe)
		for _, c := range l.Children {
			c.html(w, depth)
		}
		io.WriteString(w, "</span>")
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

func (l *Link) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := l.Children[:0]
	for _, c := range l.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	l.Children = newChildren
	return op(l)
}

// Image embeds an image.
type Image struct {
	BaseNode
	Src string
}

func (*Image) isNode() {}

func (i *Image) html(w io.Writer, depth int) {
	io.WriteString(w, "<img src=\"")
	io.WriteString(w, htmlEscape(i.Src))
	fmt.Fprintf(w, `" data-start-pos="%d" data-end-pos="%d" />`, i.Start, i.End)
}

func (i *Image) a4code(w io.Writer) {
	io.WriteString(w, "[img=")
	escapeArg(w, i.Src)
	writeByte(w, ']')
}

func (i *Image) Transform(op func(Node) (Node, error)) (Node, error) {
	return op(i)
}

// Code block.
type Code struct {
	BaseNode
	InnerStart int
	InnerEnd   int
	Value      string
}

func (*Code) isNode() {}

func (c *Code) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<pre class="a4code-block a4code-code" data-start-pos="%d" data-end-pos="%d">`, c.Start, c.End)
	fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, c.InnerStart, c.InnerEnd)
	io.WriteString(w, htmlEscape(c.Value))
	io.WriteString(w, "</span></pre>")
}

func (c *Code) a4code(w io.Writer) {
	io.WriteString(w, "[code]")
	io.WriteString(w, c.Value)
	io.WriteString(w, "[/code]")
}

func (c *Code) Transform(op func(Node) (Node, error)) (Node, error) {
	return op(c)
}

// Quote node.
type Quote struct {
	BaseNode
	Children []Node
}

func (*Quote) isNode()                {}
func (q *Quote) childrenPtr() *[]Node { return &q.Children }

func (q *Quote) html(w io.Writer, depth int) {
	colorClass := fmt.Sprintf("quote-color-%d", depth%6)
	fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quote %s" data-start-pos="%d" data-end-pos="%d">`, colorClass, q.Start, q.End)
	io.WriteString(w, "<div class=\"quote-body\">")
	for _, c := range q.Children {
		c.html(w, depth+1)
	}
	io.WriteString(w, "</div>")
	io.WriteString(w, "</blockquote>")
}

func (q *Quote) a4code(w io.Writer) {
	io.WriteString(w, "[quote")
	for _, c := range q.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

func (q *Quote) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := q.Children[:0]
	for _, c := range q.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	q.Children = newChildren
	return op(q)
}

// QuoteOf node.
type QuoteOf struct {
	BaseNode
	Name     string
	Children []Node
}

func (*QuoteOf) isNode()                {}
func (q *QuoteOf) childrenPtr() *[]Node { return &q.Children }

func (q *QuoteOf) html(w io.Writer, depth int) {
	colorClass := fmt.Sprintf("quote-color-%d", depth%6)
	fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quoteof %s" data-start-pos="%d" data-end-pos="%d"><div>Quote of `, colorClass, q.Start, q.End)
	io.WriteString(w, "<div class=\"quote-header\">Quote of ")
	io.WriteString(w, htmlEscape(q.Name))
	io.WriteString(w, ":</div>")
	io.WriteString(w, "<div class=\"quote-body\">")
	for _, c := range q.Children {
		c.html(w, depth+1)
	}
	io.WriteString(w, "</div>")
	io.WriteString(w, "</blockquote>")
}

func (q *QuoteOf) a4code(w io.Writer) {
	io.WriteString(w, "[quoteof ")
	escapeQuotedArg(w, q.Name)
	for _, c := range q.Children {
		c.a4code(w)
	}
	writeByte(w, ']')
}

func (q *QuoteOf) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := q.Children[:0]
	for _, c := range q.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	q.Children = newChildren
	return op(q)
}

// Spoiler node.
type Spoiler struct {
	BaseNode
	Children []Node
}

func (*Spoiler) isNode()                {}
func (s *Spoiler) childrenPtr() *[]Node { return &s.Children }

func (s *Spoiler) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<span class="spoiler" data-start-pos="%d" data-end-pos="%d">`, s.Start, s.End)
	for _, c := range s.Children {
		c.html(w, depth)
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

func (s *Spoiler) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := s.Children[:0]
	for _, c := range s.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	s.Children = newChildren
	return op(s)
}

// Indent node.
type Indent struct {
	BaseNode
	Children []Node
}

func (*Indent) isNode()                {}
func (i *Indent) childrenPtr() *[]Node { return &i.Children }

func (i *Indent) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<div class="a4code-block a4code-indent" data-start-pos="%d" data-end-pos="%d"><div>`, i.Start, i.End)
	for _, c := range i.Children {
		c.html(w, depth)
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

func (i *Indent) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := i.Children[:0]
	for _, c := range i.Children {
		res, err := c.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	i.Children = newChildren
	return op(i)
}

// HR node.
type HR struct {
	BaseNode
}

func (*HR) isNode() {}

func (h *HR) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<hr data-start-pos="%d" data-end-pos="%d" />`, h.Start, h.End)
}

func (*HR) a4code(w io.Writer) { io.WriteString(w, "[hr]") }

func (h *HR) Transform(op func(Node) (Node, error)) (Node, error) {
	return op(h)
}

// Custom element for unrecognised tags.
type Custom struct {
	BaseNode
	Tag      string
	Children []Node
}

func (*Custom) isNode()                {}
func (c *Custom) childrenPtr() *[]Node { return &c.Children }

func (c *Custom) html(w io.Writer, depth int) {
	fmt.Fprintf(w, `<span data-start-pos="%d" data-end-pos="%d">`, c.Start, c.End)
	io.WriteString(w, "[")
	io.WriteString(w, htmlEscape(c.Tag))
	for _, ch := range c.Children {
		ch.html(w, depth)
	}
	io.WriteString(w, "]")
	io.WriteString(w, "</span>")
}

func (c *Custom) a4code(w io.Writer) {
	writeByte(w, '[')
	io.WriteString(w, c.Tag)
	for _, ch := range c.Children {
		ch.a4code(w)
	}
	writeByte(w, ']')
}

func (c *Custom) Transform(op func(Node) (Node, error)) (Node, error) {
	newChildren := c.Children[:0]
	for _, ch := range c.Children {
		res, err := ch.Transform(op)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newChildren = append(newChildren, res)
		}
	}
	c.Children = newChildren
	return op(c)
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
