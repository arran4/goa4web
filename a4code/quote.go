package a4code

import (
	"bytes"
	"fmt"
	"strings"
)

// QuoteOption configures behaviour of Quote.
type QuoteOption func(*quoteOptions)

type quoteOptions struct {
	// Full splits the input into paragraphs and quotes each separately.
	Full bool
	// Trim removes leading and trailing whitespace from quoted text.
	Trim bool
}

// WithParagraphQuote enables paragraph aware quoting.
func WithParagraphQuote() QuoteOption { return func(o *quoteOptions) { o.Full = true } }

// WithTrimSpace removes surrounding whitespace from the quoted text.
func WithTrimSpace() QuoteOption { return func(o *quoteOptions) { o.Trim = true } }

// WithFullQuote is a backward-compatible alias for paragraph-aware quoting.
// Deprecated: use WithParagraphQuote instead.
func WithFullQuote() QuoteOption { return WithParagraphQuote() }

// QuoteText wraps the provided text in quote markup for the given user.
// Behaviour can be customised through QuoteOption values.
func QuoteText(username, text string, opts ...QuoteOption) string {
	var o quoteOptions
	for _, opt := range opts {
		opt(&o)
	}
	if o.Full {
		return fullQuoteOf(username, text, o.Trim)
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

func fullQuoteOf(username, text string, trim bool) string {
	if trim {
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
				if strings.TrimSpace(s) != "" && !isQuoteOfQuote(s) {
					quote.WriteString(quoteOfText(username, s, trim))
					quote.WriteString("\n\n")
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
	if strings.TrimSpace(s) != "" && !isQuoteOfQuote(s) {
		quote.WriteString(quoteOfText(username, s, trim))
	}
	return quote.String()
}

func isQuoteOfQuote(s string) bool {
	s = strings.TrimSpace(s)
	if !isQuoteBlock(s) {
		return false
	}
	root, err := ParseString(s)
	if err != nil || len(root.Children) != 1 {
		return false
	}
	q, ok := root.Children[0].(*QuoteOf)
	if !ok {
		return false
	}
	hasQuote := false
	hasContent := false
	for _, child := range q.Children {
		switch n := child.(type) {
		case *QuoteOf:
			hasQuote = true
		case *Text:
			if strings.TrimSpace(n.Value) != "" {
				hasContent = true
			}
		default:
			// Image, Code, etc.
			hasContent = true
		}
	}
	return hasQuote && !hasContent
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
