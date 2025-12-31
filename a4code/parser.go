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

func streamImpl(r io.Reader, yield func(Node, int) bool) {
	br := bufio.NewReader(r)
	var stack []parent
	var buf bytes.Buffer

	flush := func() bool {
		if buf.Len() == 0 {
			return true
		}
		t := &Text{Value: buf.String()}
		buf.Reset()
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), t)
		}
		return yield(t, len(stack)+1)
	}

	for {
		ch, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				if !flush() {
				}
				for len(stack) > 0 {
					n := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
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
			if !flush() {
				return
			}
			var e error
			stack, e = parseCommand(br, stack, len(stack)+1, yield)
			if e != nil {
				return
			}
		case ']':
			if !flush() {
				return
			}
			if len(stack) > 0 {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if len(stack) > 0 {
					*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n.(Node))
				}
				if !yield(n.(Node), len(stack)+1) {
					return
				}
			}
		case '\\':
			next, err := br.ReadByte()
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
			buf.WriteByte(ch)
		}
	}
}

// Parse reads markup from r and returns the root node.
func Parse(r io.Reader) (*Root, error) {
	nodes := slices.Collect(Stream(r, WithDepth(1)))
	return &Root{Children: nodes}, nil
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

func parseCommand(r *bufio.Reader, stack []parent, depth int, yield func(Node, int) bool) ([]parent, error) {
	cmd, err := getNext(r, true)
	if err != nil && err != io.EOF {
		return stack, err
	}
	switch strings.ToLower(cmd) {
	case "*", "b", "bold":
		stack = append(stack, &Bold{})
	case "/", "i", "italic":
		stack = append(stack, &Italic{})
	case "_", "u", "underline":
		stack = append(stack, &Underline{})
	case "^", "p", "power", "sup":
		stack = append(stack, &Sup{})
	case ".", "s", "sub":
		stack = append(stack, &Sub{})
	case "img", "image":
		skipArgPrefix(r)
		raw, err := getNext(r, false)
		if err != nil && err != io.EOF {
			return stack, err
		}
		n := &Image{Src: raw}
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n)
		}
		yield(n, depth)
	case "a", "link", "url":
		skipArgPrefix(r)
		raw, err := getNext(r, false)
		if err != nil && err != io.EOF {
			return stack, err
		}
		stack = append(stack, &Link{Href: raw})
	case "code":
		skipArgPrefix(r)
		raw, err := directOutput(r, "[/code]", "code]")
		if err != nil {
			return stack, err
		}
		n := &Code{Value: raw}
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n)
		}
		yield(n, depth)
	case "quoteof":
		skipArgPrefix(r)
		name, err := getNext(r, false)
		if err != nil && err != io.EOF {
			return stack, err
		}
		stack = append(stack, &QuoteOf{Name: name})
	case "quote", "q":
		stack = append(stack, &Quote{})
	case "spoiler", "sp":
		stack = append(stack, &Spoiler{})
	case "indent":
		stack = append(stack, &Indent{})
	case "hr":
		n := &HR{}
		if len(stack) > 0 {
			*stack[len(stack)-1].childrenPtr() = append(*stack[len(stack)-1].childrenPtr(), n)
		}
		yield(n, depth)
	default:
		stack = append(stack, &Custom{Tag: cmd})
	}
	return stack, nil
}

func getNext(r *bufio.Reader, endAtEqual bool) (string, error) {
	var result bytes.Buffer
	for {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return result.String(), io.EOF
			}
			return "", err
		}
		switch ch {
		case '\n', ']', '[', ' ', '\r':
			if err := r.UnreadByte(); err != nil {
				return "", err
			}
			return result.String(), nil
		case '=':
			if endAtEqual {
				if err := r.UnreadByte(); err != nil {
					return "", err
				}
				return result.String(), nil
			}
			result.WriteByte(ch)
		case '\\':
			next, err := r.ReadByte()
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

func skipArgPrefix(r *bufio.Reader) {
	if ch, err := r.ReadByte(); err == nil {
		if ch != '=' && ch != ' ' {
			r.UnreadByte()
		}
	}
}

func directOutput(r *bufio.Reader, terminators ...string) (string, error) {
	lens := make([]int, len(terminators))
	for i, t := range terminators {
		lens[i] = len(t)
	}
	var buf bytes.Buffer
	for {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return buf.String(), nil
			}
			return "", err
		}
		switch ch {
		case '\\':
			next, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					buf.WriteByte('\\')
					return buf.String(), nil
				}
				return "", err
			}
			buf.WriteByte(next)
		default:
			buf.WriteByte(ch)
			for idx, term := range terminators {
				if buf.Len() >= lens[idx] && strings.EqualFold(term, buf.String()[buf.Len()-lens[idx]:]) {
					out := buf.Bytes()[:buf.Len()-lens[idx]]
					return string(out), nil
				}
			}
		}
	}
}
