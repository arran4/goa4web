package common

import (
	"bytes"
	"fmt"
)

// ProcessCommentFullQuote constructs quoted markup with the full
// contents of the comment.
func ProcessCommentFullQuote(username, text string) string {
	var out bytes.Buffer
	var quote bytes.Buffer
	var it, bc, nlc int
	for it < len(text) {
		switch text[it] {
		case ']':
			bc--
		case '[':
			bc++
		case '\\':
			if it+1 < len(text) {
				if text[it+1] == '[' || text[it+1] == ']' {
					out.WriteByte(text[it+1])
					it++
				}
			}
		case '\n':
			if bc == 0 && nlc == 1 {
				quote.WriteString(ProcessCommentQuote(username, out.String()))
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
	quote.WriteString(ProcessCommentQuote(username, out.String()))
	return quote.String()
}

// ProcessCommentQuote wraps the given text in a quoting tag referencing the user.
func ProcessCommentQuote(username, text string) string {
	return fmt.Sprintf("[quoteof \"%s\" %s]\n", username, text)
}
