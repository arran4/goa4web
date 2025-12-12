package main

import (
	"bytes"
	"errors"
	"flag"
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

func TestPrintFlagsHelp(t *testing.T) {
	fs := newFlagSet("print-test")
	fs.String("config", "", "config file path")
	fs.Bool("verbose", false, "enable verbose output")

	buf := &bytes.Buffer{}
	fs.SetOutput(buf)

	err := fs.Parse([]string{"-help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected ErrHelp parsing -help, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "print-test flags:") {
		t.Fatalf("flag group title missing from help output: %q", output)
	}
	for _, want := range []string{"config file path", "enable verbose output"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing flag description %q in output: %q", want, output)
		}
	}
}
