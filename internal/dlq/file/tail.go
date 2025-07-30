package file

import (
	"os"
	"strings"
)

// Tail returns the last n lines of the file at path.
func Tail(path string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if n < len(lines) {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}
