package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteUsageWithGroups(t *testing.T) {
	r := &rootCmd{}
	r.fs = newFlagSet("prog")
	r.fs.String("config", "", "config file")

	var buf bytes.Buffer
	if err := executeUsage(&buf, "root_usage.txt", r); err != nil {
		t.Fatalf("executeUsage: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Global Flags") {
		t.Errorf("missing global flags: %s", out)
	}
}
