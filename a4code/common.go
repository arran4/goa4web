package a4code

import (
	"bytes"
	"io"
)

// ScannerInterface abstracts the byte reading methods used by consumeCodeBlock.
// This allows both bufio.Reader (in a4code2html) and the custom scanner (in parser.go) to be used.
type ScannerInterface interface {
	ReadByte() (byte, error)
}

// ConsumeCodeBlock consumes content bytes until the terminator ']' is found at the top level.
// It supports escaping characters with backslash.
// It returns the content string and error.
func ConsumeCodeBlock(s ScannerInterface) (string, error) {
	var buf bytes.Buffer
	const terminator = "]"
	const termLen = len(terminator)

	for {
		ch, err := s.ReadByte()
		if err != nil {
			if err == io.EOF {
				return buf.String(), nil
			}
			return "", err
		}

		if ch == '\\' {
			next, err := s.ReadByte()
			if err != nil {
				if err == io.EOF {
					buf.WriteByte('\\')
					return buf.String(), nil
				}
				return "", err
			}
			// Unescape: consume backslash, write next char
			buf.WriteByte(next)
			continue
		}

		buf.WriteByte(ch)

		if ch == ']' {
			// Found terminator "]" at top level
			// Remove the terminator from the buffer
			res := buf.String()
			res = res[:len(res)-termLen]
			return res, nil
		}
	}
}
