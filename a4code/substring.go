package a4code

import (
	"bytes"
	"strings"
)

func substring(s string, start, end int) string {
	if start < 0 || end < start {
		return ""
	}

	type tag struct {
		name  string
		open  string
		close string
	}

	// helper: write closers in reverse order
	writeClosers := func(b *bytes.Buffer, stack []tag) {
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i].close != "" {
				b.WriteString(stack[i].close)
			}
		}
	}

	// find the next ']' accounting for escapes
	findClose := func(str string, idx int) int {
		for j := idx; j < len(str); j++ {
			if str[j] == '\\' {
				j++
				continue
			}
			if str[j] == ']' {
				return j
			}
		}
		return -1
	}

	// parse tag content into a tag struct and whether it’s an opener/closer
	parseTag := func(full string) (tg tag, isOpen, isClose bool) {
		content := strings.TrimSpace(full[1 : len(full)-1]) // between [ and ]
		lower := strings.ToLower(content)
		// Determine name and args
		name := lower
		if sp := strings.IndexAny(lower, " ="); sp != -1 {
			name = lower[:sp]
		}
		if strings.HasPrefix(name, "/") {
			isClose = true
			name = strings.TrimPrefix(name, "/")
			tg = tag{name: name, open: "", close: "[" + "/" + name + "]"}
			return
		}
		// Recognize common container tags; others are treated as non-visible
		switch name {
		case "b", "i", "u", "sup", "sub", "quote", "quoteof", "spoiler", "indent", "a":
			tg = tag{name: name, open: full, close: "[" + "/" + name + "]"}
			isOpen = true
			return
		default:
			// Non-container or unknown tag; treat as non-visible (no stack effect)
			return
		}
	}

	var (
		out       bytes.Buffer
		stack     []tag
		visible   int
		inSegment bool
	)

	for i := 0; i < len(s) && visible < end; {
		// handle escapes for visible text
		if s[i] == '\\' && i+1 < len(s) {
			ch := s[i+1]
			if visible >= start && visible < end {
				if !inSegment {
					for _, t := range stack {
						out.WriteString(t.open)
					}
					inSegment = true
				}
				out.WriteByte(ch)
			}
			visible++
			i += 2
			continue
		}

		if s[i] == '[' {
			j := findClose(s, i+1)
			if j == -1 { // no closing, treat as literal
				if visible >= start && visible < end {
					if !inSegment {
						for _, t := range stack {
							out.WriteString(t.open)
						}
						inSegment = true
					}
					out.WriteByte(s[i])
				}
				visible++
				i++
				continue
			}

			// close current visible run before changing tag state
			if inSegment {
				writeClosers(&out, stack)
				inSegment = false
			}

			full := s[i : j+1]
			tg, isOpen, isClose := parseTag(full)
			if isOpen {
				stack = append(stack, tg)
			} else if isClose {
				// count closing tag as consuming one visible unit if inside window
				if visible >= start && visible < end {
					visible++
				}
				// pop until matching name (if any)
				for k := len(stack) - 1; k >= 0; k-- {
					if stack[k].name == tg.name {
						stack = stack[:k]
						break
					}
				}
			}
			i = j + 1
			continue
		}

		// visible character
		if visible >= start && visible < end {
			if !inSegment {
				for _, t := range stack {
					out.WriteString(t.open)
				}
				inSegment = true
			}
			out.WriteByte(s[i])
		}
		visible++
		i++
	}

	if inSegment {
		writeClosers(&out, stack)
	}

	return strings.TrimSpace(out.String())
}

// normaliseSimpleBB converts very simple paired BBCode tags used in tests
// into the internal single-bracket form that the parser understands.
// Currently handles only bold: [b]...[/b] -> [b...]
func normaliseSimpleBB(in string) string {
	if len(in) == 0 {
		return in
	}
	var out bytes.Buffer
	for i := 0; i < len(in); {
		// preserve escapes
		if in[i] == '\\' && i+1 < len(in) {
			out.WriteByte(in[i])
			out.WriteByte(in[i+1])
			i += 2
			continue
		}
		if strings.HasPrefix(in[i:], "[b]") {
			out.WriteString("[b")
			i += 3
			continue
		}
		if strings.HasPrefix(in[i:], "[/b]") {
			out.WriteByte(']')
			i += 4
			continue
		}
		out.WriteByte(in[i])
		i++
	}
	return out.String()
}

// openTag returns the opening tag for a node in bracket syntax, e.g. [b], [i], [a=...]
func openTag(n Node) string {
	switch t := n.(type) {
	case *Bold:
		return "[b]"
	case *Italic:
		return "[i]"
	case *Underline:
		return "[u]"
	case *Sup:
		return "[sup]"
	case *Sub:
		return "[sub]"
	case *Link:
		var b bytes.Buffer
		b.WriteString("[a=")
		escapeArg(&b, t.Href)
		b.WriteString("]")
		return b.String()
	case *Quote:
		return "[quote]"
	case *QuoteOf:
		var b bytes.Buffer
		b.WriteString("[quoteof ")
		escapeArg(&b, t.Name)
		b.WriteString("]")
		return b.String()
	case *Spoiler:
		return "[spoiler]"
	case *Indent:
		return "[indent]"
	case *Custom:
		var b bytes.Buffer
		b.WriteByte('[')
		b.WriteString(t.Tag)
		b.WriteByte(']')
		return b.String()
	default:
		return ""
	}
}

// closeTag returns the closing tag for a node in bracket syntax, e.g. [/b], [/a]
// For nodes that are not containers or do not require explicit closing, returns an empty string.
func closeTag(n Node) string {
	switch n.(type) {
	case *Bold:
		return "[/b]"
	case *Italic:
		return "[/i]"
	case *Underline:
		return "[/u]"
	case *Sup:
		return "[/sup]"
	case *Sub:
		return "[/sub]"
	case *Link:
		return "[/a]"
	case *Quote:
		return "[/quote]"
	case *QuoteOf:
		return "[/quoteof]"
	case *Spoiler:
		return "[/spoiler]"
	case *Indent:
		return "[/indent]"
	case *Custom:
		// do our best and assume symmetrical closing
		return "[/" + strings.TrimPrefix(strings.TrimSpace(nTag(n)), "/") + "]"
	default:
		return ""
	}
}

// nTag returns the raw tag name for supported nodes, or empty.
func nTag(n Node) string {
	switch t := n.(type) {
	case *Bold:
		return "b"
	case *Italic:
		return "i"
	case *Underline:
		return "u"
	case *Sup:
		return "sup"
	case *Sub:
		return "sub"
	case *Link:
		return "a"
	case *Quote:
		return "quote"
	case *QuoteOf:
		return "quoteof"
	case *Spoiler:
		return "spoiler"
	case *Indent:
		return "indent"
	case *Custom:
		return t.Tag
	default:
		return ""
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
