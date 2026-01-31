package a4code

import (
	"bytes"
	"fmt"
	"strings"
)

// QuoteOption configures behaviour of Quote.
type QuoteOption any

type quoteOptions struct {
	// Full splits the input into paragraphs and quotes each separately.
	Full bool
	// Trim removes leading and trailing whitespace from quoted text.
	Trim bool
	// RestrictedQuoteDepth is the depth at which quotes are removed.
	RestrictedQuoteDepth *int
	// TruncatedQuoteDepth is the depth at which quote content is removed.
	TruncatedQuoteDepth *int
}

// RestrictedQuoteDepth is the depth at which quotes are removed.
type RestrictedQuoteDepth int

// TruncatedQuoteDepth is the depth at which quote content is removed.
type TruncatedQuoteDepth int

// WithParagraphQuote enables paragraph aware quoting.
func WithParagraphQuote() QuoteOption { return func(o *quoteOptions) { o.Full = true } }

// WithTrimSpace removes surrounding whitespace from the quoted text.
func WithTrimSpace() QuoteOption { return func(o *quoteOptions) { o.Trim = true } }

// WithRestrictedQuoteDepth sets the depth at which quotes are removed.
func WithRestrictedQuoteDepth(depth int) QuoteOption {
	return RestrictedQuoteDepth(depth)
}

// WithTruncatedQuoteDepth sets the depth at which quote content is removed.
func WithTruncatedQuoteDepth(depth int) QuoteOption {
	return TruncatedQuoteDepth(depth)
}

// WithFullQuote is a backward-compatible alias for paragraph-aware quoting.
// Deprecated: use WithParagraphQuote instead.
func WithFullQuote() QuoteOption { return WithParagraphQuote() }

// QuoteText wraps the provided text in quote markup for the given user.
// Behaviour can be customised through QuoteOption values.
func QuoteText(username, text string, opts ...QuoteOption) string {
	var o quoteOptions
	for _, opt := range opts {
		switch v := opt.(type) {
		case func(*quoteOptions):
			v(&o)
		case RestrictedQuoteDepth:
			val := int(v)
			o.RestrictedQuoteDepth = &val
		case TruncatedQuoteDepth:
			val := int(v)
			o.TruncatedQuoteDepth = &val
		}
	}
	if o.Full {
		return fullQuoteOf(username, text, o)
	}
	return quoteOfText(username, text, o.Trim)
}

func quoteOfText(username, text string, trim bool) string {
	if trim {
		text = strings.TrimSpace(text)
	}
	return fmt.Sprintf("[quoteof \"%s\" %s]\n", escapeUsername(username), text)
}

func escapeUsername(u string) string {
	var b bytes.Buffer
	for i := 0; i < len(u); i++ {
		switch u[i] {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteByte(u[i])
		}
	}
	return b.String()
}

func fullQuoteOf(username, text string, opts quoteOptions) string {
	if opts.Trim {
		text = strings.TrimSpace(text)
	}
	var out bytes.Buffer
	var quote bytes.Buffer
	var it, bc, nlc int
	for it < len(text) {
		switch text[it] {
		case ']':
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			bc--
			out.WriteByte(text[it])
		case '[':
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			bc++
			out.WriteByte(text[it])
		case '\\':
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			if it+1 < len(text) {
				if text[it+1] == '[' || text[it+1] == ']' {
					out.WriteByte(text[it+1])
					it++
				}
			}
		case '\n':
			if bc <= 0 && nlc == 1 {
				s := out.String()
				if processed, include := processQuoteBlock(s, opts); include {
					quote.WriteString(quoteOfText(username, processed, opts.Trim))
					quote.WriteString("\n\n\n")
				}
				out.Reset()
			}
			nlc++
			it++
			continue
		case '\r':
			it++
			continue
		case ' ': // fallthrough
			fallthrough
		default:
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			out.WriteByte(text[it])
		}
		it++
	}
	s := out.String()
	if processed, include := processQuoteBlock(s, opts); include {
		quote.WriteString(quoteOfText(username, processed, opts.Trim))
	}
	return quote.String()
}

func processQuoteBlock(s string, opts quoteOptions) (string, bool) {
	sTrim := strings.TrimSpace(s)
	if sTrim == "" {
		return "", false
	}

	if !isQuoteBlock(sTrim) {
		return s, true
	}

	root, err := ParseString(sTrim)
	if err != nil || len(root.Children) != 1 {
		return s, true
	}
	q, ok := root.Children[0].(*QuoteOf)
	if !ok {
		return s, true
	}

	// Logic for RestrictedQuoteDepth
	if opts.RestrictedQuoteDepth != nil {
		if isPureQuote(q) {
			depth := getPureQuoteDepth(q)
			if depth > *opts.RestrictedQuoteDepth {
				return "", false
			}
		}
	} else {
		// Default behavior: filter if it is a pure quote block (matching original isQuoteOfQuote)
		if isPureQuote(q) {
			return "", false
		}
	}

	// Logic for TruncatedQuoteDepth
	if opts.TruncatedQuoteDepth != nil {
		truncateQuotes(root, 0, *opts.TruncatedQuoteDepth)
		return nodeToString(root), true
	}

	return s, true
}

func isPureQuote(node Node) bool {
	children := nodeChildren(node)
	if len(children) == 0 {
		return false
	}

	hasQuote := false
	for _, child := range children {
		switch child.(type) {
		case *QuoteOf:
			hasQuote = true
		case *Text:
			// Check for non-empty text
			if t, ok := child.(*Text); ok && strings.TrimSpace(t.Value) != "" {
				return false
			}
		default:
			return false // Images, code, etc count as content
		}
	}
	return hasQuote
}

func getPureQuoteDepth(node Node) int {
	max := 0
	for _, child := range nodeChildren(node) {
		if q, ok := child.(*QuoteOf); ok {
			d := 1
			if isPureQuote(q) {
				d += getPureQuoteDepth(q)
			}
			if d > max {
				max = d
			}
		}
	}
	return max
}

func nodeChildren(n Node) []Node {
	switch v := n.(type) {
	case *Root:
		return v.Children
	case *QuoteOf:
		return v.Children
	default:
		return nil
	}
}

func truncateQuotes(node Node, currentDepth int, limit int) {
	children := nodeChildren(node)
	for _, child := range children {
		if q, ok := child.(*QuoteOf); ok {
			childDepth := currentDepth + 1
			if childDepth > limit {
				q.Children = nil
			} else {
				truncateQuotes(q, childDepth, limit)
			}
		}
	}
}

// Need a helper to write AST back to string.
func nodeToString(n Node) string {
	var b bytes.Buffer
	n.a4code(&b)
	return b.String()
}

func isQuoteBlock(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(strings.ToLower(s), "[quoteof") {
		return false
	}
	// Verify it's a single block by balancing brackets
	bc := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			bc++
		case ']':
			bc--
			if bc == 0 {
				// If we closed the first tag, it must be the end of the string
				// (ignoring trailing whitespace is handled by TrimSpace above)
				return i == len(s)-1
			}
		}
	}
	return false
}

func isQuoteOf(s string) bool {
	return isQuoteBlock(s)
}

func Substring(s string, start, end int) string {
	return substring(s, start, end)
}
