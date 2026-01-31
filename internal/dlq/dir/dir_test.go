package dir

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDLQRecord(t *testing.T) {
	dir := t.TempDir()
	dlq := &DLQ{Dir: dir}

	msg := "test message"
	if err := dlq.Record(context.Background(), msg); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	content, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "Date: ") {
		t.Fatalf("expected Date header, got: %q", s)
	}
	if !strings.Contains(s, "Message-ID: ") {
		t.Fatalf("expected Message-ID header")
	}

	// Verify body is separated by double newline
	parts := strings.SplitN(s, "\n\n", 2)
	if len(parts) != 2 {
		t.Fatalf("expected headers and body separated by double newline")
	}

	expectedBody := msg + "\n"
	if parts[1] != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, parts[1])
	}

	recs, err := List(dir, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("List returned %d records", len(recs))
	}
	if recs[0].Message != strings.TrimSpace(msg) {
		t.Fatalf("List returned message %q, expected %q", recs[0].Message, strings.TrimSpace(msg))
	}
}

func TestListLegacy(t *testing.T) {
	dir := t.TempDir()
	msg := "legacy message"

	// Create legacy file (no headers)
	path := filepath.Join(dir, "legacy.txt")
	if err := os.WriteFile(path, []byte(msg), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	recs, err := List(dir, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0].Message != msg {
		t.Fatalf("expected message %q, got %q", msg, recs[0].Message)
	}
}
