package file

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDLQRecord(t *testing.T) {
	orig := appendFile
	var name string
	var data []byte
	appendFile = func(p string, d []byte) error {
		name = p
		data = append([]byte(nil), d...)
		return nil
	}
	defer func() { appendFile = orig }()

	dlq := &DLQ{Path: "test.log"}
	if err := dlq.Record(context.Background(), "hello"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if name != "test.log" {
		t.Fatalf("path=%q", name)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Fatalf("lines=%d", len(lines))
	}
	if lines[0] != fileSeparator || lines[3] != fileSeparator {
		t.Fatalf("separators wrong: %v", lines)
	}
	if _, err := time.Parse(time.RFC3339, lines[1]); err != nil {
		t.Fatalf("timestamp: %v", err)
	}
	if lines[2] != "hello" {
		t.Fatalf("message=%q", lines[2])
	}
}
