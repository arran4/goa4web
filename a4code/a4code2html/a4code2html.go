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
}

// WithTOC enables or disables table-of-contents generation when passed to New.
type WithTOC bool

// New returns a configured A4code2html converter. Optional arguments may set
// the output CodeType, enable table of contents generation or provide a custom
// ImageURLMapper. A *bufio.Reader, io.Reader or io.Writer may be supplied to
// configure the input or output streams.
func New(opts ...interface{}) *A4code2html {
	c := &A4code2html{
		CodeType: CTHTML,
		w:        new(bytes.Buffer),
	}
	for _, o := range opts {
		switch v := o.(type) {
		case CodeType:
			c.CodeType = v
		case func(tag, val string) string:
			c.ImageURLMapper = v
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
	c.r = nil
	c.w = nil
	c.err = nil
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
	lens := make([]int, len(terminators))
	for i, t := range terminators {
		lens[i] = len(t)
	}
	var buf bytes.Buffer
	for {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				_, werr := w.Write(buf.Bytes())
				return werr
			}
			return err
		}
		switch ch {
		case '\\':
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
		case '<', '>', '&':
			buf.WriteString(c.Escape(ch))
		default:
			buf.WriteByte(ch)
			for idx, term := range terminators {
				if buf.Len() >= lens[idx] && strings.EqualFold(term, buf.String()[buf.Len()-lens[idx]:]) {
					out := buf.Bytes()[:buf.Len()-lens[idx]]
					if _, err := w.Write(out); err != nil {
						return err
					}
					return nil
				}
			}
		}
	}
}

func (a *A4code2html) acommReader(r *bufio.Reader, w io.Writer) error {
	cmd, err := a.getNextReader(r, true)
	if err != nil && err != io.EOF {
		return err
	}
	_, err = a.readWhiteSpace(r)
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
		if a.ImageURLMapper != nil {
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
			if _, err := a.getNextReader(r, false); err != nil && err != io.EOF {
				return err
			}
		default:
			raw, err := a.getNextReader(r, false)
			if err != nil && err != io.EOF {
				return err
			}
			if a.ImageURLMapper != nil {
				raw = a.ImageURLMapper("a", raw)
			}
			safe, ok := SanitizeURL(raw)
			if ok {
				if _, err := io.WriteString(w, "<a href=\""+safe+"\" target=\"_BLANK\">"); err != nil {
					return err
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
			if _, err := io.WriteString(w, "<pre class=\"a4code-block a4code-code\">"); err != nil {
				return err
			}
			if err := a.directOutputReader(r, w, "[/code]", "code]"); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "</pre>"); err != nil {
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
			if _, err := io.WriteString(w, fmt.Sprintf("<blockquote class=\"a4code-block a4code-quoteof\"><div>Quote of %s:</div>", name)); err != nil {
				return err
			}
			a.stack = append(a.stack, "</blockquote>")
		}
	case "quote", "q":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<blockquote class=\"a4code-block a4code-quote\">"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</blockquote>")
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
			}
		case '<', '>', '\n', '&':
			if _, err := io.WriteString(w, c.Escape(ch)); err != nil {
				return err
			}
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
		default:
			if _, err := w.Write([]byte{ch}); err != nil {
				return err
			}
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
