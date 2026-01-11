package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRoleListSQL(t *testing.T) {
	t.Parallel()

	root := &rootCmd{fs: newFlagSet("prog")}
	var buf bytes.Buffer
	root.fs.SetOutput(&buf)

	parent := &roleCmd{rootCmd: root, fs: newFlagSet("role")}
	parent.fs.SetOutput(&buf)

	cmd, err := parseRoleListCmd(parent, []string{"sql"})
	if err != nil {
		t.Fatalf("parseRoleListCmd: %v", err)
	}

	cmd.fs.SetOutput(&buf)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"content-writer", "faq-reader", "labeler", "moderator", "moderator-sectional", "news-reader", "news-writer", "private-forum-user"}
	got := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(got) != len(want) {
		t.Fatalf("unexpected line count: got %d want %d (%v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("line %d mismatch: got %q want %q", i, got[i], w)
		}
	}
}

func TestRoleListNames(t *testing.T) {
	t.Parallel()

	root := &rootCmd{fs: newFlagSet("prog")}
	var buf bytes.Buffer
	root.fs.SetOutput(&buf)

	parent := &roleCmd{rootCmd: root, fs: newFlagSet("role")}
	parent.fs.SetOutput(&buf)

	cmd, err := parseRoleListCmd(parent, []string{"names"})
	if err != nil {
		t.Fatalf("parseRoleListCmd: %v", err)
	}

	cmd.fs.SetOutput(&buf)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"content writer", "faq reader", "labeler", "moderator", "moderator-sectional", "news reader", "news writer", "private forum user"}
	got := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(got) != len(want) {
		t.Fatalf("unexpected line count: got %d want %d (%v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("line %d mismatch: got %q want %q", i, got[i], w)
		}
	}
}
