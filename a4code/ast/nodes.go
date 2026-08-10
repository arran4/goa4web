package ast

import (
	"strings"
)

// Node represents a parsed element of markup.
type Node interface {
	String() string
	isNode()
	Transform(op func(Node) (Node, error)) (Node, error)
	SetPos(start, end int)
	GetPos() (int, int)
	GetParent() Node
	SetParent(Node)
}

type BaseNode struct {
	Start  int
	End    int
	Parent Node
}

func (n *BaseNode) SetPos(start, end int) {
	n.Start = start
	n.End = end
}

func (n *BaseNode) GetPos() (int, int) {
	return n.Start, n.End
}

func (n *BaseNode) GetParent() Node {
	return n.Parent
}

func (n *BaseNode) SetParent(p Node) {
	n.Parent = p
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

func (*Root) isNode()      {}
func (*Root) isBlockType() {}

func (r *Root) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(r, op)
}

func (r *Root) childrenPtr() *[]Node { return &r.Children }
func (r *Root) AddChild(n Node)      { n.SetParent(r); r.Children = append(r.Children, n) }
func (r *Root) GetChildren() []Node  { return r.Children }

func (r *Root) String() string {
	return joinChildren(r.Children)
}

// Text contains plain text content.
type Text struct {
	BaseNode
	Value string
}

func (*Text) isNode()       {}
func (*Text) isInlineType() {}

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
func (*Bold) isInlineType()          {}
func (b *Bold) childrenPtr() *[]Node { return &b.Children }
func (b *Bold) AddChild(n Node)      { n.SetParent(b); b.Children = append(b.Children, n) }
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
func (*Italic) isInlineType()          {}
func (i *Italic) childrenPtr() *[]Node { return &i.Children }
func (i *Italic) AddChild(n Node)      { n.SetParent(i); i.Children = append(i.Children, n) }
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
func (*Underline) isInlineType()          {}
func (u *Underline) childrenPtr() *[]Node { return &u.Children }
func (u *Underline) AddChild(n Node)      { n.SetParent(u); u.Children = append(u.Children, n) }
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
func (*Sup) isInlineType()          {}
func (s *Sup) childrenPtr() *[]Node { return &s.Children }
func (s *Sup) AddChild(n Node)      { n.SetParent(s); s.Children = append(s.Children, n) }
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
func (*Sub) isInlineType()          {}
func (s *Sub) childrenPtr() *[]Node { return &s.Children }
func (s *Sub) AddChild(n Node)      { n.SetParent(s); s.Children = append(s.Children, n) }
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
}

func (*Link) isNode()                {}
func (*Link) isInlineType()          {}
func (l *Link) Blockable() bool      { return true }
func (l *Link) childrenPtr() *[]Node { return &l.Children }
func (l *Link) AddChild(n Node)      { n.SetParent(l); l.Children = append(l.Children, n) }
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

func (*Image) isNode()       {}
func (*Image) isInlineType() {}

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

func (*Code) isNode()           {}
func (*Code) isBlockType()      {}
func (c *Code) Inlinable() bool { return !strings.Contains(c.Value, "\n") }

func (c *Code) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(c, op)
}

func (c *Code) String() string {
	return "[code " + c.Value + "]"
}

// CodeIn block with language specification.
type CodeIn struct {
	BaseNode
	Language   string
	InnerStart int
	InnerEnd   int
	Value      string
}

func (*CodeIn) isNode()           {}
func (*CodeIn) isBlockType()      {}
func (c *CodeIn) Inlinable() bool { return !strings.Contains(c.Value, "\n") }

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

func (*Quote) isNode()      {}
func (*Quote) isBlockType() {}
func (q *Quote) Inlinable() bool {
	hasNewline := false
	var walk func(Node)
	walk = func(n Node) {
		if t, ok := n.(*Text); ok && strings.Contains(t.Value, "\n") {
			hasNewline = true
			return
		}
		if p, ok := n.(parent); ok {
			for _, child := range *p.childrenPtr() {
				walk(child)
			}
		}
	}
	walk(q)
	return !hasNewline
}
func (q *Quote) childrenPtr() *[]Node { return &q.Children }
func (q *Quote) AddChild(n Node)      { n.SetParent(q); q.Children = append(q.Children, n) }
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
func (*QuoteOf) isBlockType()           {}
func (q *QuoteOf) childrenPtr() *[]Node { return &q.Children }
func (q *QuoteOf) AddChild(n Node)      { n.SetParent(q); q.Children = append(q.Children, n) }
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
func (*Spoiler) isInlineType()          {}
func (s *Spoiler) childrenPtr() *[]Node { return &s.Children }
func (s *Spoiler) AddChild(n Node)      { n.SetParent(s); s.Children = append(s.Children, n) }
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
func (*Indent) isBlockType()           {}
func (i *Indent) childrenPtr() *[]Node { return &i.Children }
func (i *Indent) AddChild(n Node)      { n.SetParent(i); i.Children = append(i.Children, n) }
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

func (*HR) isNode()      {}
func (*HR) isBlockType() {}

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
func (*Custom) isInlineType()          {}
func (c *Custom) childrenPtr() *[]Node { return &c.Children }
func (c *Custom) AddChild(n Node)      { n.SetParent(c); c.Children = append(c.Children, n) }
func (c *Custom) GetChildren() []Node  { return c.Children }

func (c *Custom) Transform(op func(Node) (Node, error)) (Node, error) {
	return transformChildren(c, op)
}

func (c *Custom) String() string {
	return "[" + c.Tag + joinChildren(c.Children) + "]"
}

// Block indicates a node that defaults to block-level rendering.
type Block interface {
	Node
	isBlockType()
}

// BlockWithInlinable indicates a block node that can optionally be rendered inline.
type BlockWithInlinable interface {
	Block
	Inlinable() bool
}

// Inline indicates a node that defaults to inline rendering.
type Inline interface {
	Node
	isInlineType()
}

// InlineWithBlockable indicates an inline node that can optionally be rendered as a block.
type InlineWithBlockable interface {
	Inline
	Blockable() bool
}

// IsBlockNode computes whether a node behaves as a block based on its context and type.
func IsBlockNode(n Node) bool {
	if _, ok := n.(*Root); ok {
		return true
	}

	p := n.GetParent()
	if p == nil {
		if _, ok := n.(Block); ok {
			return true
		}
		return false
	}

	// It's a block context if parent is block context
	isContextBlock := IsBlockNode(p)

	// Get siblings
	var siblings []Node
	if container, ok := p.(Container); ok {
		siblings = container.GetChildren()
	} else {
		// Cannot determine siblings, fallback
		if _, ok := n.(Block); ok {
			return true
		}
		return false
	}

	// Find this node's index
	idx := -1
	for i, s := range siblings {
		if s == n {
			idx = i
			break
		}
	}
	if idx == -1 {
		// Not found in parent? Fallback
		if _, ok := n.(Block); ok {
			return true
		}
		return false
	}

	// A line is bounded by start of siblings, or a newline text, or a hard block.
	lineStartIdx := 0
	for i := idx - 1; i >= 0; i-- {
		s := siblings[i]
		if txt, ok := s.(*Text); ok {
			if strings.Contains(txt.Value, "\n") || strings.Contains(txt.Value, "\r") {
				lineStartIdx = i + 1
				break
			}
		} else if _, ok := s.(Block); ok {
			if _, inlinable := s.(BlockWithInlinable); !inlinable {
				// strict block breaks the line
				lineStartIdx = i + 1
				break
			}
		}
	}

	// Also find the line end index
	lineEndIdx := len(siblings)
	for i := idx + 1; i < len(siblings); i++ {
		s := siblings[i]
		if txt, ok := s.(*Text); ok {
			if strings.Contains(txt.Value, "\n") || strings.Contains(txt.Value, "\r") {
				lineEndIdx = i
				break
			}
		} else if _, ok := s.(Block); ok {
			if _, inlinable := s.(BlockWithInlinable); !inlinable {
				lineEndIdx = i
				break
			}
		}
	}

	// Scan the line for strict inline content
	hasStrictInline := false
	for i := lineStartIdx; i < lineEndIdx; i++ {
		s := siblings[i]
		if txt, ok := s.(*Text); ok {
			val := txt.Value
			if i == lineStartIdx {
				val = strings.TrimLeft(val, " \t\n\r")
			}
			if i == lineEndIdx-1 {
				val = strings.TrimRight(val, " \t\n\r")
			} else {
				if strings.TrimSpace(val) == "" {
					val = ""
				}
			}
			if len(val) > 0 {
				hasStrictInline = true
				break
			}
		} else if _, ok := s.(InlineWithBlockable); ok {
			// can be either
		} else if _, ok := s.(BlockWithInlinable); ok {
			// can be either
		} else if _, ok := s.(Inline); ok {
			hasStrictInline = true
			break
		}
	}

	// Now apply the resolution
	if ib, ok := n.(InlineWithBlockable); ok {
		if hasStrictInline {
			return false
		} else if ib.Blockable() && isContextBlock {
			return true
		}
		return false
	} else if bi, ok := n.(BlockWithInlinable); ok {
		if hasStrictInline && bi.Inlinable() {
			return false
		} else {
			return true
		}
	} else if _, ok := n.(Block); ok {
		return true
	}

	return false
}
