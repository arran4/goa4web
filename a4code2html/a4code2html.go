// Package a4code2html converts a small markup language into HTML or
// alternative formats.
package a4code2html

import (
	"bytes"
	"fmt"
	"html"
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

func (c *A4code2html) getNext(endAtEqual bool) string {
	result := new(bytes.Buffer)
	var ch byte
	loop := true

	for loop && len(c.input) > 0 {
		ch = c.input[0]
		c.input = c.input[1:]

		switch ch {
		case '\n', ']', '[', ' ', '\r':
			loop = false
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
	command := strings.ToLower(a.getNext(true))
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
			a.getNext(false)
		default:
			raw := a.getNext(false)
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
			a.output.WriteString(fmt.Sprintf("<table width=90%% align=center bgcolor=lightgreen><tr><th>Quote of %s: <tr><td>", a.getNext(false)))
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
			a.output.WriteString("<hr>")
			a.stack = append(a.stack, "/>")
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

func (c *A4code2html) Process() {
	c.nextcomm()
	for len(c.stack) > 0 {
		last := c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]
		c.output.WriteString(last)
	}
}
