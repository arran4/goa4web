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
	return fmt.Sprintf("[quoteof \"%s\" %s]\n", username, text)
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
			bc--
			out.WriteByte(text[it])
		case '[':
			bc++
			out.WriteByte(text[it])
		case '\\':
			if it+1 < len(text) {
				if text[it+1] == '[' || text[it+1] == ']' {
					out.WriteByte(text[it+1])
					it++
				}
			}
		case '\n':
			if bc == 0 && nlc == 1 {
				quote.WriteString(quoteOfText(username, out.String(), trim))
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
	quote.WriteString(quoteOfText(username, out.String(), trim))
	return quote.String()
}

func Substring(s string, start, end int) string {
	return substring(s, start, end)
}
