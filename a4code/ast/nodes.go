package ast

import (
	"fmt"
	"strings"
)

// Node represents a parsed element of markup.
type Node interface {
	fmt.Stringer
	isNode()
	Transform(op func(Node) (Node, error)) (Node, error)
	SetPos(start, end int)
	GetPos() (int, int)
}

type BaseNode struct {
	Start   int
	End     int
	IsBlock bool
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

// Container represents a node that can hold other nodes.
type Container interface {
	Node
	AddChild(Node)
	GetChildren() []Node
}

func transformChildren(n Node, op func(Node) (Node, error)) (Node, error) {
	if p, ok := n.(parent); ok {
		children := p.childrenPtr()
		newChildren := (*children)[:0]
		for _, c := range *children {
			res, err := c.Transform(op)
			if err != nil {
				return nil, err
			}
			if res != nil {
				newChildren = append(newChildren, res)
			}
		}
		*children = newChildren
	}
	return op(n)
}

func joinChildren(children []Node) string {
	var b strings.Builder
	for _, c := range children {
		b.WriteString(c.String())
	}
	return b.String()
}

// Root is the top level node of a document.
type Root struct {
	BaseNode
	Children []Node
}

func (*Root) isNode() {}

func (r *Root) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(r, op)
}

func (r *Root) childrenPtr() *[]Node { return &r.Children }
func (r *Root) AddChild(n Node)      { r.Children = append(r.Children, n) }
func (r *Root) GetChildren() []Node  { return r.Children }

func (r *Root) String() string {
	return joinChildren(r.Children)
}

// Text contains plain text content.
type Text struct {
	BaseNode
	Value string
}

func (*Text) isNode() {}

func (t *Text) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(t, op)
}

func (t *Text) String() string {
	return t.Value
}

// Bold text.
type Bold struct {
	BaseNode
	Children []Node
}

func (*Bold) isNode()                {}
func (b *Bold) childrenPtr() *[]Node { return &b.Children }
func (b *Bold) AddChild(n Node)      { b.Children = append(b.Children, n) }
func (b *Bold) GetChildren() []Node  { return b.Children }

func (b *Bold) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(b, op)
}

func (b *Bold) String() string {
	return "[b" + joinChildren(b.Children) + "]" // Assuming implicit close or handled by parser for roundtrip, but String() is often debug or raw content representation.
	// Actually, based on previous feedback "children?", the user implies recursively printing children.
	// Simple concatenation for now.
}

// Italic text.
type Italic struct {
	BaseNode
	Children []Node
}

func (*Italic) isNode()                {}
func (i *Italic) childrenPtr() *[]Node { return &i.Children }
func (i *Italic) AddChild(n Node)      { i.Children = append(i.Children, n) }
func (i *Italic) GetChildren() []Node  { return i.Children }

func (i *Italic) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(i, op)
}

func (i *Italic) String() string {
	return "[i" + joinChildren(i.Children) + "]"
}

// Underline text.
type Underline struct {
	BaseNode
	Children []Node
}

func (*Underline) isNode()                {}
func (u *Underline) childrenPtr() *[]Node { return &u.Children }
func (u *Underline) AddChild(n Node)      { u.Children = append(u.Children, n) }
func (u *Underline) GetChildren() []Node  { return u.Children }

func (u *Underline) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(u, op)
}

func (u *Underline) String() string {
	return "[u" + joinChildren(u.Children) + "]"
}

// Superscript text.
type Sup struct {
	BaseNode
	Children []Node
}

func (*Sup) isNode()                {}
func (s *Sup) childrenPtr() *[]Node { return &s.Children }
func (s *Sup) AddChild(n Node)      { s.Children = append(s.Children, n) }
func (s *Sup) GetChildren() []Node  { return s.Children }

func (s *Sup) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(s, op)
}

func (s *Sup) String() string {
	return "[sup" + joinChildren(s.Children) + "]"
}

// Subscript text.
type Sub struct {
	BaseNode
	Children []Node
}

func (*Sub) isNode()                {}
func (s *Sub) childrenPtr() *[]Node { return &s.Children }
func (s *Sub) AddChild(n Node)      { s.Children = append(s.Children, n) }
func (s *Sub) GetChildren() []Node  { return s.Children }

func (s *Sub) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(s, op)
}

func (s *Sub) String() string {
	return "[sub" + joinChildren(s.Children) + "]"
}

// Link to a URL.
type Link struct {
	BaseNode
	Href     string
	Children []Node
	IsBlock  bool
}

func (*Link) isNode()                {}
func (l *Link) childrenPtr() *[]Node { return &l.Children }
func (l *Link) AddChild(n Node)      { l.Children = append(l.Children, n) }
func (l *Link) GetChildren() []Node  { return l.Children }

func (l *Link) IsImmediateClose() bool {
	return len(l.Children) == 0
}

func (l *Link) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(l, op)
}

func (l *Link) String() string {
	return "[link " + l.Href + joinChildren(l.Children) + "]"
}

// Image embeds an image.
type Image struct {
	BaseNode
	Src string
}

func (*Image) isNode() {}

func (i *Image) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(i, op)
}

func (i *Image) String() string {
	return "[img=" + i.Src + "]"
}

// Code block.
type Code struct {
	BaseNode
	InnerStart int
	InnerEnd   int
	Value      string
}

func (*Code) isNode() {}

func (c *Code) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(c, op)
}

func (c *Code) String() string {
	return "[code]" + c.Value + "[/code]"
}

// CodeIn block with language specification.
type CodeIn struct {
	BaseNode
	Language   string
	InnerStart int
	InnerEnd   int
	Value      string
}

func (*CodeIn) isNode() {}

func (c *CodeIn) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(c, op)
}

func (c *CodeIn) String() string {
	return "[codein \"" + c.Language + "\" " + c.Value + "]"
}

// Quote node.
type Quote struct {
	BaseNode
	Children []Node
}

func (*Quote) isNode()                {}
func (q *Quote) childrenPtr() *[]Node { return &q.Children }
func (q *Quote) AddChild(n Node)      { q.Children = append(q.Children, n) }
func (q *Quote) GetChildren() []Node  { return q.Children }

func (q *Quote) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(q, op)
}

func (q *Quote) String() string {
	return "[quote" + joinChildren(q.Children) + "]"
}

// QuoteOf node.
type QuoteOf struct {
	BaseNode
	Name     string
	Children []Node
}

func (*QuoteOf) isNode()                {}
func (q *QuoteOf) childrenPtr() *[]Node { return &q.Children }
func (q *QuoteOf) AddChild(n Node)      { q.Children = append(q.Children, n) }
func (q *QuoteOf) GetChildren() []Node  { return q.Children }

func (q *QuoteOf) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(q, op)
}

func (q *QuoteOf) String() string {
	return "[quoteof " + q.Name + joinChildren(q.Children) + "]"
}

// Spoiler node.
type Spoiler struct {
	BaseNode
	Children []Node
}

func (*Spoiler) isNode()                {}
func (s *Spoiler) childrenPtr() *[]Node { return &s.Children }
func (s *Spoiler) AddChild(n Node)      { s.Children = append(s.Children, n) }
func (s *Spoiler) GetChildren() []Node  { return s.Children }

func (s *Spoiler) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(s, op)
}

func (s *Spoiler) String() string {
	return "[spoiler" + joinChildren(s.Children) + "]"
}

// Indent node.
type Indent struct {
	BaseNode
	Children []Node
}

func (*Indent) isNode()                {}
func (i *Indent) childrenPtr() *[]Node { return &i.Children }
func (i *Indent) AddChild(n Node)      { i.Children = append(i.Children, n) }
func (i *Indent) GetChildren() []Node  { return i.Children }

func (i *Indent) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(i, op)
}

func (i *Indent) String() string {
	return "[indent" + joinChildren(i.Children) + "]"
}

// HR node.
type HR struct {
	BaseNode
}

func (*HR) isNode() {}

func (h *HR) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(h, op)
}

func (h *HR) String() string {
	return "[hr]"
}

// Custom element for unrecognised tags.
type Custom struct {
	BaseNode
	Tag      string
	Children []Node
}

func (*Custom) isNode()                {}
func (c *Custom) childrenPtr() *[]Node { return &c.Children }
func (c *Custom) AddChild(n Node)      { c.Children = append(c.Children, n) }
func (c *Custom) GetChildren() []Node  { return c.Children }

func (c *Custom) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(c, op)
}

func (c *Custom) String() string {
	return "[" + c.Tag + joinChildren(c.Children) + "]"
}
