package main

import (
	"bytes"
	"fmt"
	"strings"
)

type codetype int

const (
	ct_html codetype = iota
	ct_tableOfContents
	ct_tagstrip
	ct_wordsonly
)

type A4code2html struct {
	input    string
	output   bytes.Buffer
	codeType codetype
	makeTC   bool
	stack    []string
}

func NewA4Code2HTML() *A4code2html {
	return &A4code2html{
		codeType: ct_html,
	}
}

func (c *A4code2html) clear() {
	c.input = ""
	c.output.Reset()
	c.stack = nil
}

func (c *A4code2html) Escape(ch byte) string {
	if c.codeType == ct_wordsonly {
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
		switch c.codeType {
		case ct_tagstrip:
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

	for loop {
		ch = c.input[0]
		c.input = c.input[1:]

		switch ch {
		case '\n', ']', '[', ' ', '\r', '=':
			loop = false
		case '\\':
			ch = c.input[0]
			c.input = c.input[1:]
			if ch != ' ' && ch != '[' && ch != ']' && ch != '=' && ch != '\\' {
				result.WriteByte('\\')
			}
		default:
			result.WriteByte(ch)
		}
	}

	return result.String()
}

func (c *A4code2html) directOutput(terminator string) {
	lensomething := len(terminator)
	var last string

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
			if i := len(c.output.Bytes()) - lensomething; i >= 0 {
				last = c.output.String()[i:]
				if strings.EqualFold(terminator, last) {
					c.output.Truncate(i)
					return
				}
			}
		}
	}
}

func (a *A4code2html) acomm() int {
	command := a.getNext(true)
	switch command {
	case "*", "b", "bold":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<strong>")
			a.stack = append(a.stack, "</strong>")
		}
	case "/", "i", "italic":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<i>")
			a.stack = append(a.stack, "</i>")
		}
	case "_", "u", "underline":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<u>")
			a.stack = append(a.stack, "</u>")
		}
	case "^", "p", "power", "sup":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<sup>")
			a.stack = append(a.stack, "</sup>")
		}
	case ".", "s", "sub":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<sub>")
			a.stack = append(a.stack, "</sub>")
		}
	case "img", "image":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<img src=\"")
			a.stack = append(a.stack, "\" />")
		}
	case "a", "link", "url":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
			a.getNext(false)
		default:
			// TODO make URL safe
			a.output.WriteString("<a href=\"" + a.getNext(false) + "\" target=\"_BLANK\">")
			a.stack = append(a.stack, "</a>")
		}
	case "code":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<table width=90% align=center bgcolor=lightblue><tr><th>Code: <tr><td><pre>")
			a.output.WriteString("</pre></table>")
		}
	case "quoteof":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString(fmt.Sprintf("<table width=90%% align=center bgcolor=lightgreen><tr><th>Quote of %s: <tr><td>", a.getNext(false)))
			a.stack = append(a.stack, "</table>")
		}
	case "quote", "q":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<table width=90% align=center bgcolor=lightgreen><tr><th>Quote: <tr><td>")
			a.stack = append(a.stack, "</table>")
		}
	case "indent":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
		default:
			a.output.WriteString("<table width=90% align=center><tr><td>")
			a.stack = append(a.stack, "</table>")
		}
	case "hr":
		switch a.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
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
			if ch != ' ' && ch != '[' && ch != ']' && ch != '=' && ch != '\\' {
				c.output.WriteByte('\\')
			}
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
