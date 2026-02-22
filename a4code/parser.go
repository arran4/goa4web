package a4code

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/arran4/goa4web/a4code/ast"
)

type streamOptions struct {
	maxDepth int // nodes deeper than this level are skipped; -1 yields all
}

// StreamOption configures Stream behaviour.
type StreamOption func(*streamOptions)

// WithDepth limits yielded nodes to the specified depth, where 0 is top level
// and -1 yields all nodes.
func WithDepth(d int) StreamOption { return func(o *streamOptions) { o.maxDepth = d } }

// WithAllNodes yields every node encountered while parsing.
func WithAllNodes() StreamOption { return func(o *streamOptions) { o.maxDepth = -1 } }

// Stream parses markup from r and yields nodes according to the provided options.
func Stream(r io.Reader, opts ...StreamOption) iter.Seq[ast.Node] {
	o := streamOptions{maxDepth: -1}
	for _, op := range opts {
		op(&o)
	}

	return func(yield func(ast.Node) bool) {
		internalYield := func(n ast.Node, level int) bool {
			if o.maxDepth == -1 || level <= o.maxDepth {
				yield(n)
			}
			return true
		}
		streamImpl(r, internalYield)
	}
}

type scanner struct {
	r   *bufio.Reader
	pos int // raw byte position (unused for AST now, but kept for low level logic if needed)
}

func (s *scanner) ReadByte() (byte, error) {
	b, err := s.r.ReadByte()
	if err == nil {
		s.pos++
	}
	return b, err
}

func (s *scanner) UnreadByte() error {
	err := s.r.UnreadByte()
	if err == nil {
		s.pos--
	}
	return err
}

func (s *scanner) Peek() (byte, error) {
	b, err := s.r.Peek(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func isBlockContext(n ast.Node) bool {
	if _, ok := n.(*ast.Root); ok {
		return true
	}
	// Check specific types that have IsBlock field
	// Actually, the AST nodes embed BaseNode, but the interface ast.Node doesn't expose it directly except via getter?
	// But we are casting to specific types usually.
	// Let's use type switch and access the field.
	switch t := n.(type) {
	case *ast.Quote:
		return t.IsBlock
	case *ast.QuoteOf:
		return t.IsBlock // Always true
	case *ast.Spoiler:
		return t.IsBlock
	case *ast.Indent:
		return t.IsBlock // Always true
	case *ast.Link:
		return t.IsBlock
	case *ast.Code:
		return t.IsBlock
	}
	return false
}

func isBlockTag(tag string) bool {
	switch strings.ToLower(tag) {
	case "quote", "quoteof", "spoiler", "indent":
		return true
	}
	return false
}

func updateBlockStatus(children []ast.Node, newChild ast.Node, isContextBlock bool) {
	if len(children) > 0 {
		prev := children[len(children)-1]
		if l, ok := prev.(*ast.Link); ok && l.IsBlock {
			// Check if newChild starts with newline or is a block element
			startsNewline := false

			if isBlockContext(newChild) {
				startsNewline = true
			} else if txt, ok := newChild.(*ast.Text); ok {
				if strings.HasPrefix(txt.Value, "\n") {
					startsNewline = true
				}
			}

			if !startsNewline {
				l.IsBlock = false
			}
		}
	}

	if isContextBlock {
		if l, ok := newChild.(*ast.Link); ok {
			// Check previous sibling, skipping whitespace
			prevIsNewline := false

			// Start checking from the last child
			idx := len(children) - 1
			for idx >= 0 {
				lastChild := children[idx]

				if isBlockContext(lastChild) {
					prevIsNewline = true
					break
				} else if txt, ok := lastChild.(*ast.Text); ok {
					if strings.HasSuffix(txt.Value, "\n") {
						prevIsNewline = true
						break
					}
					if strings.TrimSpace(txt.Value) != "" {
						// Found non-whitespace text that doesn't end in newline
						prevIsNewline = false
						break
					}
					// If strictly whitespace, continue looking back
				} else {
					// Other inline element (e.g. bold, italic, image)
					prevIsNewline = false
					break
				}
				idx--
			}

			// If we exhausted children without finding content (or children was empty), it's start of block
			if idx < 0 {
				prevIsNewline = true
			}

			if prevIsNewline {
				l.IsBlock = true
			} else {
				l.IsBlock = false
			}
		}
	}
}

func streamImpl(r io.Reader, yield func(ast.Node, int) bool) {
	br := bufio.NewReader(r)
	s := &scanner{r: br, pos: 0}
	var stack []ast.Container
	var buf bytes.Buffer

	// visiblePos tracks the byte offset in the "visible" text (content of Text and Code nodes).
	// Tags themselves do not advance this counter.
	visiblePos := 0

	textStart := -1
	lastChar := byte('\n')

	flush := func(offset int) bool {
		if buf.Len() == 0 {
			textStart = -1
			return true
		}
		t := &ast.Text{Value: buf.String()}
		// Text node range is current visiblePos to visiblePos + len
		start := visiblePos
		end := visiblePos + len(t.Value)
		t.SetPos(start, end)
		visiblePos = end

		textStart = -1
		buf.Reset()
		if len(stack) > 0 {
			p := stack[len(stack)-1]
			children := p.GetChildren()
			updateBlockStatus(children, t, isBlockContext(p.(ast.Node)))
			p.AddChild(t)
		}
		return yield(t, len(stack)+1)
	}

	for {
		ch, err := s.ReadByte()
		if err != nil {
			if err == io.EOF {
				if !flush(0) {
				}
				for len(stack) > 0 {
					n := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					if nNode, ok := n.(ast.Node); ok {
						nNode.SetPos(nNode.GetPos())
						start, _ := nNode.GetPos()
						nNode.SetPos(start, visiblePos)
					}

					if len(stack) > 0 {
						p := stack[len(stack)-1]
						children := p.GetChildren()
						updateBlockStatus(children, n, isBlockContext(p.(ast.Node)))
						p.AddChild(n)
					}
					if !yield(n, len(stack)+1) {
					}
				}
				return
			}
			return
		}
		switch ch {
		case '[':
			if !flush(1) {
				return
			}
			var e error
			// startPos for tag is current visiblePos
			startPos := visiblePos
			stack, visiblePos, e = parseCommand(s, stack, len(stack)+1, yield, startPos, visiblePos, lastChar)
			lastChar = ']' // Assume command ended with ]
			if e != nil {
				// We should probably return the error or handle it, but streamImpl signature doesn't return error.
				// For now, we return (stop parsing).
				return
			}
		case ']':
			if !flush(1) {
				return
			}
			if len(stack) > 0 {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if nNode, ok := n.(ast.Node); ok {
					start, _ := nNode.GetPos()
					nNode.SetPos(start, visiblePos)

					// Determine IsBlock for closed node (Quote, etc)
					switch t := nNode.(type) {
					case *ast.Quote:
						if t.IsBlock {
							next, err := s.Peek()
							if err == io.EOF || (err == nil && (next == '\n' || next == '\r')) {
								// Kept as block
							} else {
								t.IsBlock = false
							}
						}
					case *ast.QuoteOf:
						t.IsBlock = true
					case *ast.Indent:
						t.IsBlock = true
					}
				}

				if len(stack) > 0 {
					p := stack[len(stack)-1]
					children := p.GetChildren()
					updateBlockStatus(children, n, isBlockContext(p.(ast.Node)))
					p.AddChild(n)
				}
				if !yield(n, len(stack)+1) {
					return
				}
			}
			lastChar = ']'
		case '\\':
			if textStart == -1 {
				textStart = s.pos - 1
			}
			next, err := s.ReadByte()
			if err != nil {
				if err == io.EOF {
					buf.WriteByte('\\')
					continue
				}
				return
			}
			switch next {
			case ' ', '[', ']', '=', '\\', '*', '/', '_':
				buf.WriteByte(next)
			default:
				buf.WriteByte('\\')
				buf.WriteByte(next)
			}
			lastChar = next
		default:
			if textStart == -1 {
				textStart = s.pos - 1
			}
			buf.WriteByte(ch)
			lastChar = ch
		}
	}
}

// Parse reads markup from r and returns the root node.
func Parse(r io.Reader) (*ast.Root, error) {
	var nodes []ast.Node
	root := &ast.Root{} // Temporary root to check context logic

	for n := range Stream(r, WithDepth(1)) {
		// Update block status for Root children
		updateBlockStatus(nodes, n, true) // Root is always block context
		nodes = append(nodes, n)
	}

	root.Children = nodes
	// Calculate root range based on children
	if len(nodes) > 0 {
		start, _ := nodes[0].GetPos()
		_, end := nodes[len(nodes)-1].GetPos()
		// Since we track visiblePos now, start/end are consistent.
		root.SetPos(start, end)
	} else {
		root.SetPos(0, 0)
	}
	return root, nil
}

// ParseString parses markup from s and returns the root node.
func ParseString(s string) (*ast.Root, error) {
	return Parse(strings.NewReader(s))
}

// ParseNodesReader parses r and returns only the top-level nodes.
func ParseNodesReader(r io.Reader) ([]ast.Node, error) {
	var nodes []ast.Node
	for n := range Stream(r, WithDepth(1)) {
		updateBlockStatus(nodes, n, true) // Treat as root context for block logic
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// ParseNodes parses s and returns only the top-level nodes.
func ParseNodes(s string) ([]ast.Node, error) {
	return ParseNodesReader(strings.NewReader(s))
}

func parseCommand(s *scanner, stack []ast.Container, depth int, yield func(ast.Node, int) bool, startPos int, visiblePos int, lastChar byte) ([]ast.Container, int, error) {
	cmd, err := GetNext(s, true)
	if err != nil && err != io.EOF {
		return stack, visiblePos, err
	}

	// Error on closing tags
	if strings.HasPrefix(cmd, "/") {
		return stack, visiblePos, fmt.Errorf("closing tags like [%s] are not supported in a4code; use lisp-style nesting [tag content]", cmd)
	}

	createNode := func(n ast.Container) {
		skipArgPrefix(s)      // Consume any whitespace separator between tag and content
		n.SetPos(startPos, 0) // End will be set when popped

		parentIsBlock := true
		if len(stack) > 0 {
			parentIsBlock = isBlockContext(stack[len(stack)-1].(ast.Node))
		}

		// Determine initial IsBlock status
		isBlockStart := (lastChar == '\n' || lastChar == '\r' || startPos == 0) && parentIsBlock

		// Set IsBlock on the node if possible
		switch t := n.(type) {
		case *ast.Quote:
			t.IsBlock = isBlockStart
		case *ast.QuoteOf:
			t.IsBlock = true // QuoteOf is always block
		case *ast.Link:
			t.IsBlock = isBlockStart
		case *ast.Indent:
			t.IsBlock = true
		case *ast.Spoiler:
			// Spoiler usually inline?
			t.IsBlock = false
		}

		stack = append(stack, n)
	}

	switch strings.ToLower(cmd) {
	case "*", "b", "bold":
		createNode(&ast.Bold{})
	case "/", "i", "italic":
		createNode(&ast.Italic{})
	case "_", "u", "underline":
		createNode(&ast.Underline{})
	case "^", "p", "power", "sup":
		createNode(&ast.Sup{})
	case ".", "s", "sub":
		createNode(&ast.Sub{})
	case "img", "image":
		skipArgPrefix(s)
		raw, err := GetNext(s, false)
		if err != nil && err != io.EOF {
			return stack, visiblePos, err
		}
		if ch, err := s.ReadByte(); err == nil {
			if ch != ']' {
				s.UnreadByte()
			}
		}
		n := &ast.Image{Src: raw}
		n.SetPos(startPos, visiblePos) // Self-closing, 0-width in visible space
		if len(stack) > 0 {
			p := stack[len(stack)-1]
			children := p.GetChildren()
			updateBlockStatus(children, n, isBlockContext(p.(ast.Node)))
			p.AddChild(n)
		}
		yield(n, depth)
	case "a", "link", "url":
		skipArgPrefix(s)
		raw, err := GetNext(s, false)
		if err != nil && err != io.EOF {
			return stack, visiblePos, err
		}
		n := &ast.Link{Href: raw}
		createNode(n)
	case "code":
		skipArgPrefix(s)

		if ch, err := s.Peek(); err == nil && ch == ']' {
			s.ReadByte() // Consume ']' which starts the block in legacy syntax
		}

		// ConsumeCodeBlock consumes content bytes until terminator
		// Support [code ... ]
		startContentPos := s.pos
		raw, err := ConsumeCodeBlock(s)
		if err != nil {
			return stack, visiblePos, err
		}
		// endContentPos := s.pos - 1 // -1 for the closing bracket
		_ = startContentPos

		// raw is the content.
		contentLen := len(raw)
		innerStart := visiblePos
		innerEnd := visiblePos + contentLen

		n := &ast.Code{Value: raw, InnerStart: innerStart, InnerEnd: innerEnd}
		n.SetPos(startPos, innerEnd) // Code node includes content

		// Determine IsBlock for Code
		isBlockStart := lastChar == '\n' || lastChar == '\r' || startPos == 0
		isBlockEnd := false
		next, err := s.Peek()
		if err == io.EOF || (err == nil && (next == '\n' || next == '\r')) {
			isBlockEnd = true
		}

		n.IsBlock = isBlockStart && isBlockEnd

		visiblePos += contentLen

		if len(stack) > 0 {
			p := stack[len(stack)-1]
			children := p.GetChildren()
			updateBlockStatus(children, n, isBlockContext(p.(ast.Node)))
			p.AddChild(n)
		}
		yield(n, depth)
	case "codein":
		skipArgPrefix(s)
		language, err := GetNextArg(s)
		if err != nil && err != io.EOF {
			return stack, visiblePos, err
		}
		skipArgPrefix(s)
		// ConsumeCodeBlock consumes content bytes
		startContentPos := s.pos
		raw, err := ConsumeCodeBlock(s)
		if err != nil {
			return stack, visiblePos, err
		}
		// endContentPos := s.pos - 1
		_ = startContentPos

		// raw is the content.
		contentLen := len(raw)
		innerStart := visiblePos
		innerEnd := visiblePos + contentLen

		n := &ast.CodeIn{Language: language, Value: raw, InnerStart: innerStart, InnerEnd: innerEnd}
		n.SetPos(startPos, innerEnd) // Code node includes content

		visiblePos += contentLen

		if len(stack) > 0 {
			p := stack[len(stack)-1]
			children := p.GetChildren()
			updateBlockStatus(children, n, isBlockContext(p.(ast.Node)))
			p.AddChild(n)
		}
		yield(n, depth)
	case "quoteof":
		skipArgPrefix(s)
		name, err := GetNextArg(s)
		if err != nil && err != io.EOF {
			return stack, visiblePos, err
		}
		n := &ast.QuoteOf{Name: name}
		createNode(n)
	case "quote", "q":
		createNode(&ast.Quote{})
	case "spoiler", "sp":
		n := &ast.Spoiler{}
		createNode(n)
		n.IsBlock = lastChar == '\n' || lastChar == '\r' || startPos == 0
	case "indent":
		createNode(&ast.Indent{})
	case "hr":
		n := &ast.HR{}
		if ch, err := s.ReadByte(); err == nil {
			if ch != ']' {
				s.UnreadByte()
			}
		}
		n.SetPos(startPos, visiblePos)
		if len(stack) > 0 {
			p := stack[len(stack)-1]
			children := p.GetChildren()
			updateBlockStatus(children, n, isBlockContext(p.(ast.Node)))
			p.AddChild(n)
		}
		yield(n, depth)
	default:
		// Custom tag
		n := &ast.Custom{Tag: cmd}
		createNode(n)
	}
	return stack, visiblePos, nil
}

func skipArgPrefix(s *scanner) {
	ch, err := s.ReadByte()
	if err != nil {
		return
	}

	if ch == ' ' || ch == '=' {
		// Optional newline after space/eq
		next, err := s.ReadByte()
		if err != nil {
			return
		}
		if next == '\n' {
			return
		}
		if next == '\r' {
			if next2, err := s.ReadByte(); err == nil && next2 != '\n' {
				s.UnreadByte()
			}
			return
		}
		s.UnreadByte()
		return
	}

	if ch == '\n' {
		return
	}
	if ch == '\r' {
		if next, err := s.ReadByte(); err == nil && next != '\n' {
			s.UnreadByte()
		}
		return
	}

	s.UnreadByte()
}

