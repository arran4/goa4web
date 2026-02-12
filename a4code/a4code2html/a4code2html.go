// Package a4code2html converts a small markup language into HTML or
// alternative formats.
package a4code2html

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"strings"
)

// CodeType defines the output mode for A4code2html.
type CodeType int

const (
	// CTHTML produces standard HTML output.
	CTHTML CodeType = iota

	// CTTableOfContents outputs only the table of contents.
	CTTableOfContents

	// CTTagStrip removes all formatting tags.
	CTTagStrip

	// CTWordsOnly returns only the raw words.
	CTWordsOnly
)

// LinkProvider provides HTML rendering for links and maps image URLs.
type LinkProvider interface {
	RenderLink(url string, isBlock bool, isImmediateClose bool) (htmlOpen string, htmlClose string, consumeImmediate bool)
	MapImageURL(tag, val string) string
}

type A4code2html struct {
	r        *bufio.Reader
	w        io.Writer
	CodeType CodeType
	makeTC   bool
	stack    []string
	err      error
	// ImageURLMapper optionally maps tag URLs to fully qualified versions.
	// The first parameter provides the tag name, e.g. "img" or "a".
	ImageURLMapper func(tag, val string) string
	// UserColorMapper optionally maps a username to a CSS class for styling quotes.
	UserColorMapper func(username string) string
	quoteDepth      int
	// Provider optionally provides custom rendering for links.
	Provider    LinkProvider
	atLineStart bool
}

// WithTOC enables or disables table-of-contents generation when passed to New.
type WithTOC bool

// New returns a configured A4code2html converter. Optional arguments may set
// the output CodeType, enable table of contents generation or provide a custom
// ImageURLMapper. A *bufio.Reader, io.Reader or io.Writer may be supplied to
// configure the input or output streams.
func New(opts ...interface{}) *A4code2html {
	c := &A4code2html{
		CodeType:    CTHTML,
		w:           new(bytes.Buffer),
		atLineStart: true,
	}
	for _, o := range opts {
		switch v := o.(type) {
		case CodeType:
			c.CodeType = v
		case func(tag, val string) string:
			c.ImageURLMapper = v
		case func(string) string:
			c.UserColorMapper = v
		case LinkProvider:
			c.Provider = v
		case WithTOC:
			c.makeTC = bool(v)
		case *bufio.Reader:
			c.r = v
		case io.Reader:
			c.r = bufio.NewReader(v)
		case string:
			c.SetInput(v)
		case []byte:
			c.SetInput(string(v))
		case io.Writer:
			c.w = v
		}
	}
	return c
}

// SanitizeURL validates a hyperlink and returns a safe version.
func SanitizeURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return html.EscapeString(raw), false
	}
	switch u.Scheme {
	case "http", "https":
		return html.EscapeString(u.String()), true
	default:
		return html.EscapeString(raw), false
	}
}

func (c *A4code2html) clear() {
	c.stack = nil
	c.quoteDepth = 0
	c.r = nil
	c.w = nil
	c.err = nil
	c.atLineStart = true
}

// SetInput assigns the text to be processed.
func (c *A4code2html) SetInput(s string) {
	c.r = bufio.NewReader(strings.NewReader(s))
}

// SetReader assigns the reader supplying the markup.
func (c *A4code2html) SetReader(r io.Reader) {
	c.r = bufio.NewReader(r)
}

// SetWriter assigns the destination for rendered output.
func (c *A4code2html) SetWriter(w io.Writer) {
	c.w = w
}

func (c *A4code2html) Escape(ch byte) string {
	if c.CodeType == CTWordsOnly {
		return " "
	}
	switch ch {
	case '&':
		return "&amp;"
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	case '\n':
		switch c.CodeType {
		case CTTagStrip:
			return "\n"
		default:
			return "<br />\n"
		}
	default:
		return ""
	}
}

// getNextReader reads characters from r until it reaches a control character.
func (c *A4code2html) getNextReader(r *bufio.Reader, endAtEqual bool) (string, error) {
	result := new(bytes.Buffer)
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

func (c *A4code2html) directOutputReader(r *bufio.Reader, w io.Writer, terminators ...string) error {
	if len(terminators) == 0 {
		terminators = []string{"]"}
	}
	var buf bytes.Buffer
	depth := 0

	for {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				_, werr := w.Write(buf.Bytes())
				return werr
			}
			return err
		}

		if ch == '\\' {
			next, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					buf.WriteByte('\\')
					_, werr := w.Write(buf.Bytes())
					return werr
				}
				return err
			}
			buf.WriteByte(next)
			continue
		} else if ch == '<' || ch == '>' || ch == '&' {
			buf.WriteString(c.Escape(ch))
			continue
		}

		buf.WriteByte(ch)

		// Check multi-char terminators
        for _, t := range terminators {
            if len(t) > 1 && strings.HasSuffix(buf.String(), t) {
                // Found terminator
                out := buf.Bytes()[:buf.Len()-len(t)]
                if _, err := w.Write(out); err != nil {
                    return err
                }
                return nil
            }
        }

		switch ch {
		case '[':
			depth++
		case ']':
             // Check if "]" is a terminator
             isTerm := false
             for _, t := range terminators {
                 if t == "]" {
                     isTerm = true
                     break
                 }
             }

             if isTerm && depth == 0 {
                 // Found terminator "]" at top level
                 out := buf.Bytes()[:buf.Len()-1]
                 if _, err := w.Write(out); err != nil {
                     return err
                 }
                 return nil
             }

             if depth > 0 {
                 depth--
             }
		}
	}
}

func (a *A4code2html) peekBlockLink(r *bufio.Reader) (bool, bool) {
	// Returns (isBlock, isImmediateClose)
	limit := 4096
	p, err := r.Peek(limit)

	for i, b := range p {
		if b == ']' {
			// Check what follows
			if i+1 >= len(p) {
				if err == io.EOF {
					return true, i == 0
				}
				return false, false
			}
			next := p[i+1]
			if next == '\n' || next == '\r' {
				return true, i == 0
			}
			return false, false
		}
		if b == '\n' || b == '\r' {
			return false, false
		}
	}
	return false, false
}

func (a *A4code2html) acommReader(r *bufio.Reader, w io.Writer) error {
	wasAtLineStart := a.atLineStart
	a.atLineStart = false

	cmd, err := a.getNextReader(r, true)
	if err != nil && err != io.EOF {
		return err
	}
	_, err = a.readCommandBreak(r)
	if err != nil && err != io.EOF {
		return err
	}
	switch strings.ToLower(cmd) {
	case "*", "b", "bold":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<strong>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</strong>")
		}
	case "/", "i", "italic":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<i>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</i>")
		}
	case "_", "u", "underline":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<u>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</u>")
		}
	case "^", "p", "power", "sup":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<sup>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</sup>")
		}
	case ".", "s", "sub":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<sub>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</sub>")
		}
	case "img", "image":
		raw, err := a.getNextReader(r, false)
		if err != nil && err != io.EOF {
			return err
		}
		if a.Provider != nil {
			raw = a.Provider.MapImageURL("img", raw)
		} else if a.ImageURLMapper != nil {
			raw = a.ImageURLMapper("img", raw)
		}
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<img class=\"a4code-image\" src=\""+html.EscapeString(raw)+"\" />"); err != nil {
				return err
			}
		}
	case "a", "link", "url":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
			raw, err := a.getNextReader(r, false)
			if err != nil && err != io.EOF {
				return err
			}
			if p, err := r.Peek(1); err == nil && len(p) > 0 && p[0] == ']' {
				if _, err := io.WriteString(w, raw); err != nil {
					return err
				}
			}
		default:
			raw, err := a.getNextReader(r, false)
			if err != nil && err != io.EOF {
				return err
			}

			if a.Provider != nil {
				isBlock := false
				isImmediateClose := false
				if wasAtLineStart {
					isBlock, isImmediateClose = a.peekBlockLink(r)
				}
				if !isBlock {
					p, err := r.Peek(1)
					if err == nil && len(p) > 0 && p[0] == ']' {
						isImmediateClose = true
					}
				}
				htmlOpen, htmlClose, consumeImmediate := a.Provider.RenderLink(raw, isBlock, isImmediateClose)
				if _, err := io.WriteString(w, htmlOpen); err != nil {
					return err
				}
				if consumeImmediate {
					// Consume ] and potentially \n
					r.ReadByte() // ]
					if isBlock {
						r.ReadByte() // \n
						a.atLineStart = true
					}
				} else {
					a.stack = append(a.stack, htmlClose)
				}
				return nil
			}

			original := raw
			if a.ImageURLMapper != nil {
				raw = a.ImageURLMapper("a", raw)
			}
			safe, ok := SanitizeURL(raw)
			if ok {
				// Inline Link
				if _, err := io.WriteString(w, "<a href=\""+safe+"\" target=\"_blank\">"); err != nil {
					return err
				}

				isImmediateClose := false
				p, err := r.Peek(1)
				if err == nil && len(p) > 0 && p[0] == ']' {
					isImmediateClose = true
				} else {
					p, err := r.Peek(2)
					if err == nil && len(p) >= 2 && p[0] == ' ' && p[1] == ']' {
						if _, err := r.ReadByte(); err == nil {
							isImmediateClose = true
						}
					}
				}

				if isImmediateClose {
					// Case [link url]
					// Default legacy behavior without metadata provider: just print original
					if _, err := io.WriteString(w, html.EscapeString(original)); err != nil {
						return err
					}
				}
				a.stack = append(a.stack, "</a>")
			} else {
				if _, err := io.WriteString(w, safe); err != nil {
					return err
				}
				a.stack = append(a.stack, "")
			}
		}
	case "code":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			var buf bytes.Buffer
			if p, err := r.Peek(1); err == nil && len(p) > 0 && p[0] == ']' {
				r.ReadByte() // consume ]
			} else {
				if err := a.directOutputReader(r, &buf); err != nil {
					return err
				}
			}

			isBlock := false
			if wasAtLineStart {
				// Check if followed by newline
				p, err := r.Peek(1)
				if err == io.EOF || (err == nil && len(p) > 0 && (p[0] == '\n' || p[0] == '\r')) {
					isBlock = true
				}
			}

			if isBlock {
				if _, err := io.WriteString(w, "<div class=\"a4code-block a4code-code-wrapper\"><div class=\"code-header\">Code</div><pre class=\"a4code-code-body\">"); err != nil {
					return err
				}
				if _, err := w.Write(buf.Bytes()); err != nil {
					return err
				}
				if _, err := io.WriteString(w, "</pre></div>"); err != nil {
					return err
				}
			} else {
				if _, err := io.WriteString(w, "<code class=\"a4code-inline a4code-code\">"); err != nil {
					return err
				}
				if _, err := w.Write(buf.Bytes()); err != nil {
					return err
				}
				if _, err := io.WriteString(w, "</code>"); err != nil {
					return err
				}
			}
		}
	case "codein":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			language, err := a.getNextReader(r, false)
			if err != nil && err != io.EOF {
				return err
			}
			if len(language) >= 2 && language[0] == '"' && language[len(language)-1] == '"' {
				language = language[1 : len(language)-1]
			}
			if _, err := a.readWhiteSpace(r); err != nil && err != io.EOF {
				return err
			}
			if _, err := io.WriteString(w, fmt.Sprintf("<div class=\"a4code-block a4code-code-wrapper a4code-language-%s\"><div class=\"code-header\">Code (%s)</div><pre class=\"a4code-code-body\"><code class=\"language-%s\">", html.EscapeString(language), html.EscapeString(language), html.EscapeString(language))); err != nil {
				return err
			}
			if err := a.directOutputReader(r, w, "]"); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "</code></pre></div>"); err != nil {
				return err
			}
		}
	case "quoteof":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			name, err := a.getNextReader(r, false)
			if err != nil && err != io.EOF {
				return err
			}
			colorClass := ""
			if a.UserColorMapper != nil {
				colorClass = " " + a.UserColorMapper(name)
			}
			colorClass += fmt.Sprintf(" quote-color-%d", a.quoteDepth%6)
			if _, err := io.WriteString(w, fmt.Sprintf("<blockquote class=\"a4code-block a4code-quoteof%s\"><div class=\"quote-header\">Quote of %s:</div><div class=\"quote-body\">", colorClass, name)); err != nil {
				return err
			}
			a.quoteDepth++
			a.stack = append(a.stack, "</div></blockquote>")
		}
	case "quote", "q":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if wasAtLineStart {
				// Block quote
				colorClass := fmt.Sprintf(" quote-color-%d", a.quoteDepth%6)
				if _, err := io.WriteString(w, "<blockquote class=\"a4code-block a4code-quote"+colorClass+"\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">"); err != nil {
					return err
				}
				a.quoteDepth++
				a.stack = append(a.stack, "</div></blockquote>")
			} else {
				// Inline quote
				if _, err := io.WriteString(w, "<q class=\"a4code-inline a4code-quote\">"); err != nil {
					return err
				}
				a.stack = append(a.stack, "</q>")
			}
		}
	case "spoiler", "sp":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<span class=\"spoiler\">"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</span>")
		}
	case "indent":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<div class=\"a4code-block a4code-indent\"><div>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</div></div>")
		}
	case "hr":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<hr />"); err != nil {
				return err
			}
			a.stack = append(a.stack, "")
		}
	default:
		a.stack = append(a.stack, "")
	}
	return nil
}

func (c *A4code2html) nextcommReader(r *bufio.Reader, w io.Writer) error {
	for {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch ch {
		case '[':
			if err := c.acommReader(r, w); err != nil {
				return err
			}
		case ']':
			if len(c.stack) > 0 {
				last := c.stack[len(c.stack)-1]
				c.stack = c.stack[:len(c.stack)-1]
				if _, err := io.WriteString(w, last); err != nil {
					return err
				}
				if last == "</div></blockquote>" && c.quoteDepth > 0 {
					c.quoteDepth--
				}
			}
			c.atLineStart = false
		case '<', '>', '&':
			if _, err := io.WriteString(w, c.Escape(ch)); err != nil {
				return err
			}
			c.atLineStart = false
		case '\n':
			if _, err := io.WriteString(w, c.Escape(ch)); err != nil {
				return err
			}
			c.atLineStart = true
		case '\\':
			next, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					if _, err := w.Write([]byte{'\\'}); err != nil {
						return err
					}
					return nil
				}
				return err
			}
			if next != ' ' && next != '[' && next != ']' && next != '=' && next != '\\' && next != '*' && next != '/' && next != '_' {
				if _, err := w.Write([]byte{'\\'}); err != nil {
					return err
				}
			}
			if _, err := w.Write([]byte{next}); err != nil {
				return err
			}
			c.atLineStart = false
		default:
			if _, err := w.Write([]byte{ch}); err != nil {
				return err
			}
			c.atLineStart = false
		}
	}
}

// ProcessReader converts the markup from r and writes the result to w in a streaming fashion.
func (c *A4code2html) ProcessReader(r io.Reader, w io.Writer) error {
	c.clear()
	br := bufio.NewReader(r)
	if err := c.nextcommReader(br, w); err != nil {
		return err
	}
	for len(c.stack) > 0 {
		last := c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]
		if _, err := io.WriteString(w, last); err != nil {
			return err
		}
	}
	return nil
}

// Process converts the markup from the configured reader and returns an
// io.Reader containing the result. If a writer was provided via New or
// SetWriter, the output is written there and an empty reader is returned.
func (c *A4code2html) Process() io.Reader {
	if c.r == nil {
		return bytes.NewReader(nil)
	}
	dest := c.w
	var buf *bytes.Buffer
	if dest == nil {
		buf = new(bytes.Buffer)
		dest = buf
	} else if b, ok := dest.(*bytes.Buffer); ok {
		buf = b
	}
	if err := c.ProcessReader(c.r, dest); err != nil {
		c.err = fmt.Errorf("process reader: %w", err)
		log.Print(c.err)
	}
	if buf != nil {
		return bytes.NewReader(buf.Bytes())
	}
	return bytes.NewReader(nil)
}

// Error returns the last processing error, if any.
func (c *A4code2html) Error() error { return c.err }

func (c *A4code2html) readWhiteSpace(r *bufio.Reader) (string, error) {
	result := new(bytes.Buffer)
	for {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return result.String(), io.EOF
			}
			return result.String(), nil
		}
		switch ch {
		case '\n', ' ', '\r', '\t':
			result.WriteByte(ch)
		default:
			if err := r.UnreadByte(); err != nil {
				return "", err
			}
			return result.String(), nil
		}
	}
}

func (c *A4code2html) readCommandBreak(r *bufio.Reader) (string, error) {
	var buf bytes.Buffer
	ws, err := c.readWhiteSpace(r)
	buf.WriteString(ws)
	if err != nil {
		return buf.String(), err
	}

	ch, err := r.ReadByte()
	if err != nil {
		if err == io.EOF {
			return buf.String(), io.EOF
		}
		return buf.String(), err
	}

	if ch == '=' {
		buf.WriteByte(ch)
		ws2, err := c.readWhiteSpace(r)
		buf.WriteString(ws2)
		if err != nil {
			return buf.String(), err
		}
	} else {
		r.UnreadByte()
	}
	return buf.String(), nil
}
