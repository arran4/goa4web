package config

import (
	"bufio"
	"bytes"
	"strings"
)

// ParseEnvBytes parses key=value pairs from data.
// Lines beginning with '#' or empty lines are ignored. Anything after a '#' is treated as a comment.
func ParseEnvBytes(data []byte) map[string]string {
	vals := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i := strings.IndexByte(line, '='); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			vals[key] = val
		}
	}
	return vals
}
