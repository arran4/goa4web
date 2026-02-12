package a4code

import (
	"bufio"
	"bytes"
	"io"
	"iter"
	"slices"
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
			stack[len(stack)-1].AddChild(t)
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
						stack[len(stack)-1].AddChild(n)
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
					stack[len(stack)-1].AddChild(n)
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
	nodes := slices.Collect(Stream(r, WithDepth(1)))
	root := &ast.Root{Children: nodes}
	// Calculate root range based on children
	if len(nodes) > 0 {
		start, _ := nodes[0].GetPos()
		_, end := nodes[len(nodes)-1].GetPos()
		// Since we track visiblePos now, start/end are consistent.
		// If there are gaps (not possible if stream is continuous), this might be weird,
		// but visiblePos is continuous.
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
	return slices.Collect(Stream(r, WithDepth(1))), nil
}

// ParseNodes parses s and returns only the top-level nodes.
func ParseNodes(s string) ([]ast.Node, error) {
	return ParseNodesReader(strings.NewReader(s))
}

func parseCommand(s *scanner, stack []ast.Container, depth int, yield func(ast.Node, int) bool, startPos int, visiblePos int, lastChar byte) ([]ast.Container, int, error) {
	cmd, err := getNext(s, true)
	if err != nil && err != io.EOF {
		return stack, visiblePos, err
	}

	createNode := func(n ast.Container) {
		n.SetPos(startPos, 0) // End will be set when popped
		// Determine initial IsBlock status
		isBlockStart := lastChar == '\n' || lastChar == '\r' || startPos == 0

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
		raw, err := getNext(s, false)
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
			stack[len(stack)-1].AddChild(n)
		}
		yield(n, depth)
	case "a", "link", "url":
		skipArgPrefix(s)
		raw, err := getNext(s, false)
		if err != nil && err != io.EOF {
			return stack, visiblePos, err
		}
		n := &ast.Link{Href: raw}
		createNode(n)
	case "code":
		skipArgPrefix(s)

		// directOutput consumes content bytes
		// Enable balanced brackets support for Code
		raw, _, _, err := directOutput(s)
		if err != nil {
			return stack, visiblePos, err
		}
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
			stack[len(stack)-1].AddChild(n)
		}
		yield(n, depth)
	case "quoteof":
		skipArgPrefix(s)
		name, err := getNextArg(s)
		if err != nil && err != io.EOF {
			return stack, visiblePos, err
		}
		n := &ast.QuoteOf{Name: name}
		createNode(n)
	case "quote", "q":
		createNode(&ast.Quote{})
	case "spoiler", "sp":
		createNode(&ast.Spoiler{})
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
			stack[len(stack)-1].AddChild(n)
		}
		yield(n, depth)
	default:
		// Custom tag
		n := &ast.Custom{Tag: cmd}
		createNode(n)
	}
	return stack, visiblePos, nil
}

func getNextArg(s *scanner) (string, error) {
	ch, err := s.ReadByte()
	if err != nil {
		if err == io.EOF {
			return "", io.EOF
		}
		return "", err
	}
	if ch == '"' {
		var result bytes.Buffer
		for {
			ch, err = s.ReadByte()
			if err != nil {
				if err == io.EOF {
					return result.String(), io.EOF
				}
				return "", err
			}
			switch ch {
			case '"':
				return result.String(), nil
			case '\\':
				next, err := s.ReadByte()
				if err != nil {
					if err == io.EOF {
						result.WriteByte('\\')
						return result.String(), io.EOF
					}
					return "", err
				}
				switch next {
				case '"', ' ', '[', ']', '=', '\\', '*', '/', '_':
					result.WriteByte(next)
				default:
					result.WriteByte('\\')
					result.WriteByte(next)
				}
			default:
				result.WriteByte(ch)
			}
		}
	} else {
		if err := s.UnreadByte(); err != nil {
			return "", err
		}
		return getNext(s, false)
	}
}

func getNext(s *scanner, endAtEqual bool) (string, error) {
	var result bytes.Buffer
	for {
		ch, err := s.ReadByte()
		if err != nil {
			if err == io.EOF {
				return result.String(), io.EOF
			}
			return "", err
		}
		switch ch {
		case '\n', ']', '[', ' ', '\r':
			if err := s.UnreadByte(); err != nil {
				return "", err
			}
			return result.String(), nil
		case '=':
			if endAtEqual {
				if err := s.UnreadByte(); err != nil {
					return "", err
				}
				return result.String(), nil
			}
			result.WriteByte(ch)
		case '\\':
			next, err := s.ReadByte()
			if err != nil {
				if err == io.EOF {
					result.WriteByte('\\')
					return result.String(), io.EOF
				}
				return "", err
			}
			switch next {
			case ' ', '[', ']', '=', '\\', '*', '/', '_':
				result.WriteByte(next)
			default:
				result.WriteByte('\\')
				result.WriteByte(next)
			}
		default:
			result.WriteByte(ch)
		}
	}
}

func skipArgPrefix(s *scanner) {
	if ch, err := s.ReadByte(); err == nil {
		if ch != '=' && ch != ' ' {
			s.UnreadByte()
		}
	}
}

func directOutput(s *scanner) (string, int, int, error) {
	var buf bytes.Buffer
	startPos := s.pos
	depth := 0
	terminator := "]"
	termLen := len(terminator)

	for {
		ch, err := s.ReadByte()
		if err != nil {
			if err == io.EOF {
				return buf.String(), startPos, s.pos, nil
			}
			return "", 0, 0, err
		}
		switch ch {
		case '\\':
			next, err := s.ReadByte()
			if err != nil {
				if err == io.EOF {
					buf.WriteByte('\\')
					return buf.String(), startPos, s.pos, nil
				}
				return "", 0, 0, err
			}
			buf.WriteByte(next)
		case '[':
			buf.WriteByte(ch)
			depth++
		case ']':
			buf.WriteByte(ch)
			if depth > 0 {
				depth--
			} else if buf.Len() >= termLen && strings.EqualFold(terminator, buf.String()[buf.Len()-termLen:]) {
				// Accepted
				out := buf.Bytes()[:buf.Len()-termLen]
				endPos := s.pos - termLen
				return string(out), startPos, endPos, nil
			}
		default:
			buf.WriteByte(ch)
			if depth == 0 && buf.Len() >= termLen && strings.EqualFold(terminator, buf.String()[buf.Len()-termLen:]) {
				// Accepted
				out := buf.Bytes()[:buf.Len()-termLen]
				endPos := s.pos - termLen
				return string(out), startPos, endPos, nil
			}
		}
	}
}
