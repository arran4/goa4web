package file

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

type mockAppender struct{ fn func(string, []byte) error }

func (m mockAppender) Append(n string, d []byte) error { return m.fn(n, d) }

func TestDLQRecord(t *testing.T) {
	var name string
	var data []byte
	mock := mockAppender{func(p string, d []byte) error {
		name = p
		data = append([]byte(nil), d...)
		return nil
	}}
	dlq := &DLQ{Path: "test.log", Appender: mock}
	if err := dlq.Record(context.Background(), "hello"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if name != "test.log" {
		t.Fatalf("path=%q", name)
	}

	s := string(data)
	if !strings.HasPrefix(s, "From DLQ ") {
		t.Fatalf("expected From DLQ header, got: %q", s)
	}
	if !strings.Contains(s, "\nDate: ") {
		t.Fatalf("expected Date header")
	}
	if !strings.Contains(s, "\nMessage-ID: <") {
		t.Fatalf("expected Message-ID header")
	}
	if !strings.Contains(s, "\n\nhello\n") {
		t.Fatalf("expected body 'hello'")
	}
}

func TestListLegacy(t *testing.T) {
	// Create legacy format file
	// -----
	// <Timestamp>
	// <Message>
	// -----

	f, err := os.CreateTemp("", "dlq-legacy")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	ts := time.Now().Truncate(time.Second)
	tsStr := ts.Format(time.RFC3339)
	msg := "legacy message"

	content := fmt.Sprintf("%s\n%s\n%s\n%s\n", "-----", tsStr, msg, "-----")
	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	f.Close()

	recs, err := List(f.Name(), 10)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}

	if !recs[0].Time.Equal(ts) {
		t.Fatalf("expected time %v, got %v", ts, recs[0].Time)
	}

	if recs[0].Message != msg {
		t.Fatalf("expected message %q, got %q", msg, recs[0].Message)
	}
}
