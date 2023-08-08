package main

import (
	"bytes"
	"strings"
)

type codetype int

const (
	ct_html codetype = iota
	ct_tableOfContents
	ct_tagstrip
	ct_wordsonly
)

type a4code2html struct {
	input    string
	output   bytes.Buffer
	codeType codetype
	makeTC   bool
	stack    []string
}

func newA4Code2HTML() *a4code2html {
	return &a4code2html{
		codeType: ct_html,
	}
}

func (c *a4code2html) clear() {
	c.input = ""
	c.output.Reset()
	c.stack = nil
}

func (c *a4code2html) htmlelement(ch byte) string {
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

func (c *a4code2html) getNext(endAtEqual bool) string {
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

func (c *a4code2html) directOutput(terminator string) {
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
			c.output.WriteString(c.htmlelement(ch))
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

func (c *a4code2html) acomm() {
	command := c.getNext(true)

	switch {
	case command == "*" || command == "b" || strings.EqualFold(command, "bold"):
		switch c.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
			c.output.WriteString("<strong>")
			c.stack = append(c.stack, "</strong>")
		}
	case command == "/" || command == "i" || strings.EqualFold(command, "italic"):
		switch c.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
			c.output.WriteString("<i>")
			c.stack = append(c.stack, "</i>")
		}
	case command == "_" || command == "u" || strings.EqualFold(command, "underline"):
		switch c.codeType {
		case ct_tableOfContents:
		case ct_tagstrip, ct_wordsonly:
			c.output.WriteString("<u>")
			c.stack = append(c.stack, "</u>")
		}
	// Add more cases for other commands
	default:
		c.stack = append(c.stack, "")
	}

	// Don't forget to free the memory allocated for 'command'
}

func (c *a4code2html) nextcomm() {
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
			c.output.WriteString(c.htmlelement(ch))
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

func (c *a4code2html) process() {
	c.nextcomm()
	for len(c.stack) > 0 {
		last := c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]
		c.output.WriteString(last)
	}
}
