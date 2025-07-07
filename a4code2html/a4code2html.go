// Package a4code2html converts a small markup language into HTML or
// alternative formats.
package a4code2html

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
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
	input    string
	output   bytes.Buffer
	CodeType CodeType
	makeTC   bool
	stack    []string
}

func NewA4Code2HTML() *A4code2html {
	return &A4code2html{
		CodeType: CTHTML,
	}
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
	c.input = ""
	c.output.Reset()
	c.stack = nil
}

// SetInput assigns the text to be processed.
func (c *A4code2html) SetInput(s string) {
	c.input = s
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

func (c *A4code2html) getNext(endAtEqual bool, keepClose bool) string {
	result := new(bytes.Buffer)
	var ch byte
	loop := true

	for loop && len(c.input) > 0 {
		ch = c.input[0]
		c.input = c.input[1:]

		switch ch {
		case '\n', '[', ' ', '\r':
			loop = false
		case ']':
			loop = false
			if keepClose {
				c.input = string(ch) + c.input
			}
		case '=':
			if endAtEqual {
				loop = false
			} else {
				result.WriteByte(ch)
			}
		case '\\':
			if len(c.input) > 0 {
				ch = c.input[0]
				c.input = c.input[1:]
				switch ch {
				case ' ', '[', ']', '=', '\\', '*', '/', '_':
					result.WriteByte(ch)
				default:
					result.WriteByte('\\')
					result.WriteByte(ch)
				}
			} else {
				result.WriteByte('\\')
			}
		default:
			result.WriteByte(ch)
		}
	}

	return result.String()
}

func (c *A4code2html) directOutput(terminators ...string) {
	lens := make([]int, len(terminators))
	for i, t := range terminators {
		lens[i] = len(t)
	}
	for len(c.input) > 0 {
		ch := c.input[0]
		c.input = c.input[1:]

		switch ch {
		case '\\':
			ch = c.input[0]
			c.input = c.input[1:]
			c.output.WriteByte(ch)
		case '<', '>', '&':
			c.output.WriteString(c.Escape(ch))
		default:
			c.output.WriteByte(ch)
			for idx, term := range terminators {
				if i := len(c.output.Bytes()) - lens[idx]; i >= 0 {
					if strings.EqualFold(term, c.output.String()[i:]) {
						c.output.Truncate(i)
						return
					}
				}
			}
		}
	}
}

func (a *A4code2html) acomm() int {
	command := strings.ToLower(a.getNext(true, true))
	if command == "code" && len(a.input) > 0 && a.input[0] == ']' {
		a.input = a.input[1:]
	}
	switch command {
	case "*", "b", "bold":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<strong>")
			a.stack = append(a.stack, "</strong>")
		}
	case "/", "i", "italic":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<i>")
			a.stack = append(a.stack, "</i>")
		}
	case "_", "u", "underline":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<u>")
			a.stack = append(a.stack, "</u>")
		}
	case "^", "p", "power", "sup":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<sup>")
			a.stack = append(a.stack, "</sup>")
		}
	case ".", "s", "sub":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<sub>")
			a.stack = append(a.stack, "</sub>")
		}
	case "img", "image":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<img src=\"")
			a.stack = append(a.stack, "\" />")
		}
	case "a", "link", "url":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
			a.getNext(false, false)
		default:
			raw := a.getNext(false, false)
			safe, ok := SanitizeURL(raw)
			if ok {
				a.output.WriteString("<a href=\"" + safe + "\" target=\"_BLANK\">")
				a.stack = append(a.stack, "</a>")
			} else {
				a.output.WriteString(safe)
				a.stack = append(a.stack, "")
			}
		}
	case "code":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<table width=90% align=center bgcolor=lightblue><tr><th>Code: <tr><td><pre>")
			a.directOutput("[/code]", "code]")
			a.output.WriteString("</pre></table>")
		}
	case "quoteof":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString(fmt.Sprintf("<table width=90%% align=center bgcolor=lightgreen><tr><th>Quote of %s: <tr><td>", a.getNext(false, false)))
			a.stack = append(a.stack, "</table>")
		}
	case "quote", "q":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<table width=90% align=center bgcolor=lightgreen><tr><th>Quote: <tr><td>")
			a.stack = append(a.stack, "</table>")
		}
	case "spoiler", "sp":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<span onmouseover=\"this.style.color='#FFFFFF';\" onmouseout=\"this.style.color='#000000';\" style=\"color:#000000;background:#000000;\">")
			a.stack = append(a.stack, "</span>")
		}
	case "indent":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<table width=90% align=center><tr><td>")
			a.stack = append(a.stack, "</table>")
		}
	case "hr":
		switch a.CodeType {
		case CTTableOfContents:
		case CTTagStrip, CTWordsOnly:
		default:
			a.output.WriteString("<hr")
			a.stack = append(a.stack, " />")
		}
	default:
		a.stack = append(a.stack, "")
	}
	return 0
}

func (c *A4code2html) nextcomm() {
	for len(c.input) > 0 {
		ch := c.input[0]
		c.input = c.input[1:]

		switch ch {
		case '[':
			c.acomm()
		case ']':
			if len(c.stack) > 0 {
				last := c.stack[len(c.stack)-1]
				c.stack = c.stack[:len(c.stack)-1]
				c.output.WriteString(last)
			}
		case '<', '>', '\n', '&':
			c.output.WriteString(c.Escape(ch))
		case '\\':
			ch = c.input[0]
			c.input = c.input[1:]
			if ch != ' ' && ch != '[' && ch != ']' && ch != '=' && ch != '\\' && ch != '*' && ch != '/' && ch != '_' {
				c.output.WriteByte('\\')
			}
			fallthrough
		default:
			c.output.WriteByte(ch)
		}
	}
}

func (c *A4code2html) Output() string {
	return c.output.String()
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
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<img src=\""); err != nil {
				return err
			}
			a.stack = append(a.stack, "\" />")
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
			if _, err := io.WriteString(w, "<table width=90% align=center bgcolor=lightblue><tr><th>Code: <tr><td><pre>"); err != nil {
				return err
			}
			if err := a.directOutputReader(r, w, "[/code]", "code]"); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "</pre></table>"); err != nil {
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
			if _, err := io.WriteString(w, fmt.Sprintf("<table width=90%% align=center bgcolor=lightgreen><tr><th>Quote of %s: <tr><td>", name)); err != nil {
				return err
			}
			a.stack = append(a.stack, "</table>")
		}
	case "quote", "q":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<table width=90% align=center bgcolor=lightgreen><tr><th>Quote: <tr><td>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</table>")
		}
	case "spoiler", "sp":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<span onmouseover=\"this.style.color='#FFFFFF';\" onmouseout=\"this.style.color='#000000';\" style=\"color:#000000;background:#000000;\">"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</span>")
		}
	case "indent":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<table width=90% align=center><tr><td>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "</table>")
		}
	case "hr":
		switch a.CodeType {
		case CTTableOfContents, CTTagStrip, CTWordsOnly:
		default:
			if _, err := io.WriteString(w, "<hr>"); err != nil {
				return err
			}
			a.stack = append(a.stack, "/>")
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

func (c *A4code2html) Process() {
	c.nextcomm()
	for len(c.stack) > 0 {
		last := c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]
		c.output.WriteString(last)
	}
}
