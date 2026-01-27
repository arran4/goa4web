package a4code

import (
	"bufio"
	"bytes"
	"io"
	"iter"
	"slices"
	"strings"
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
func Stream(r io.Reader, opts ...StreamOption) iter.Seq[Node] {
	o := streamOptions{maxDepth: -1}
	for _, op := range opts {
		op(&o)
	}

	return func(yield func(Node) bool) {
		internalYield := func(n Node, level int) bool {
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
	pos int
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

func streamImpl(r io.Reader, yield func(Node, int) bool) {
	br := bufio.NewReader(r)
	s := &scanner{r: br, pos: 0}
	var stack []parent
	var buf bytes.Buffer
	textStart := -1

	flush := func(offset int) bool {
		if buf.Len() == 0 {
			textStart = -1
			return true
		}
		t := &Text{Value: buf.String()}
		t.SetPos(textStart, s.pos-offset)
		textStart = -1
		buf.Reset()
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), t)
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
					if nNode, ok := n.(Node); ok {
						nNode.SetPos(nNode.GetPos())
						// End position is end of stream if not closed properly
						// But strictly, if popped here, it means implicit close at EOF.
						// We can set End to s.pos.
						start, _ := nNode.GetPos()
						nNode.SetPos(start, s.pos)
					}
					if len(stack) > 0 {
						*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n.(Node))
					}
					if !yield(n.(Node), len(stack)+1) {
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
			startPos := s.pos - 1
			stack, e = parseCommand(s, stack, len(stack)+1, yield, startPos)
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
				if nNode, ok := n.(Node); ok {
					start, _ := nNode.GetPos()
					nNode.SetPos(start, s.pos) // Include ']' in range
				}
				if len(stack) > 0 {
					*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n.(Node))
				}
				if !yield(n.(Node), len(stack)+1) {
					return
				}
			}
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
		default:
			if textStart == -1 {
				textStart = s.pos - 1
			}
			buf.WriteByte(ch)
		}
	}
}

// Parse reads markup from r and returns the root node.
func Parse(r io.Reader) (*Root, error) {
	nodes := slices.Collect(Stream(r, WithDepth(1)))
	// Root node doesn't really have a single pos if constructed from stream of nodes
	// unless we wrap the whole stream. Here we just return list of children.
	// But Root struct has fields.
	root := &Root{Children: nodes}
	// We can set root pos if we want, but usually it's implicit 0 to EOF.
	// We don't have access to total length here easily unless we track it.
	return root, nil
}

// ParseString parses markup from s and returns the root node.
func ParseString(s string) (*Root, error) {
	return Parse(strings.NewReader(s))
}

// ParseNodesReader parses r and returns only the top-level nodes.
func ParseNodesReader(r io.Reader) ([]Node, error) {
	return slices.Collect(Stream(r, WithDepth(1))), nil
}

// ParseNodes parses s and returns only the top-level nodes.
func ParseNodes(s string) ([]Node, error) {
	return ParseNodesReader(strings.NewReader(s))
}

func parseCommand(s *scanner, stack []parent, depth int, yield func(Node, int) bool, startPos int) ([]parent, error) {
	cmd, err := getNext(s, true)
	if err != nil && err != io.EOF {
		return stack, err
	}

	createNode := func(n Node) {
		n.SetPos(startPos, 0) // End will be set when popped
		stack = append(stack, n.(parent))
	}

	switch strings.ToLower(cmd) {
	case "*", "b", "bold":
		createNode(&Bold{})
	case "/", "i", "italic":
		createNode(&Italic{})
	case "_", "u", "underline":
		createNode(&Underline{})
	case "^", "p", "power", "sup":
		createNode(&Sup{})
	case ".", "s", "sub":
		createNode(&Sub{})
	case "img", "image":
		skipArgPrefix(s)
		raw, err := getNext(s, false)
		if err != nil && err != io.EOF {
			return stack, err
		}
		if ch, err := s.ReadByte(); err == nil {
			if ch != ']' {
				s.UnreadByte()
			}
		}
		n := &Image{Src: raw}
		n.SetPos(startPos, s.pos) // Self-closing (conceptually)
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n)
		}
		yield(n, depth)
	case "a", "link", "url":
		skipArgPrefix(s)
		raw, err := getNext(s, false)
		if err != nil && err != io.EOF {
			return stack, err
		}
		n := &Link{Href: raw}
		createNode(n)
	case "code":
		skipArgPrefix(s)
		raw, innerStart, innerEnd, err := directOutput(s, "[/code]", "code]")
		if err != nil {
			return stack, err
		}
		n := &Code{Value: raw, InnerStart: innerStart, InnerEnd: innerEnd}
		n.SetPos(startPos, s.pos)
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n)
		}
		yield(n, depth)
	case "quoteof":
		skipArgPrefix(s)
		name, err := getNextArg(s)
		if err != nil && err != io.EOF {
			return stack, err
		}
		n := &QuoteOf{Name: name}
		createNode(n)
	case "quote", "q":
		createNode(&Quote{})
	case "spoiler", "sp":
		createNode(&Spoiler{})
	case "indent":
		createNode(&Indent{})
	case "hr":
		n := &HR{}
		if ch, err := s.ReadByte(); err == nil {
			if ch != ']' {
				s.UnreadByte()
			}
		}
		n.SetPos(startPos, s.pos) // Self-closing? But HR is typically [hr] so yes.
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n)
		}
		yield(n, depth)
	default:
		// Custom tag
		n := &Custom{Tag: cmd}
		createNode(n)
	}
	return stack, nil
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

func directOutput(s *scanner, terminators ...string) (string, int, int, error) {
	lens := make([]int, len(terminators))
	for i, t := range terminators {
		lens[i] = len(t)
	}
	var buf bytes.Buffer
	startPos := s.pos

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
		default:
			buf.WriteByte(ch)
			for idx, term := range terminators {
				if buf.Len() >= lens[idx] && strings.EqualFold(term, buf.String()[buf.Len()-lens[idx]:]) {
					out := buf.Bytes()[:buf.Len()-lens[idx]]
					// End position of content is current pos - length of terminator
					endPos := s.pos - lens[idx]
					return string(out), startPos, endPos, nil
				}
			}
		}
	}
}
