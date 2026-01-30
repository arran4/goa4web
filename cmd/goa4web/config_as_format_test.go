package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/configformat"
)

func TestConfigAsFormattingMatchesCLI(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	root := &rootCmd{cfg: cfg}
	parent := &configCmd{rootCmd: root, fs: newFlagSet("config")}

	t.Run("as-cli", func(t *testing.T) {
		cmd, err := parseConfigAsCmd(parent, "as-cli", []string{})
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		output := captureStdout(t, func() {
			if err := cmd.asCLI(); err != nil {
				t.Fatalf("run: %v", err)
			}
		})
		expected, err := configformat.FormatAsCLI(cfg, root.ConfigFile)
		if err != nil {
			t.Fatalf("format: %v", err)
		}
		if output != expected {
			t.Fatalf("output mismatch\nexpected: %q\ngot: %q", expected, output)
		}
	})

	t.Run("as-env-file-extended", func(t *testing.T) {
		cmd, err := parseConfigAsCmd(parent, "as-env-file", []string{"--extended"})
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		output := captureStdout(t, func() {
			if err := cmd.asEnvFile(); err != nil {
				t.Fatalf("run: %v", err)
			}
		})
		expected, err := configformat.FormatAsEnvFile(cfg, root.ConfigFile, root.dbReg, configformat.AsOptions{Extended: true})
		if err != nil {
			t.Fatalf("format: %v", err)
		}
		if output != expected {
			t.Fatalf("output mismatch\nexpected: %q\ngot: %q", expected, output)
		}
	})

	t.Run("as-env", func(t *testing.T) {
		cmd, err := parseConfigAsCmd(parent, "as-env", []string{})
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		output := captureStdout(t, func() {
			if err := cmd.asEnv(); err != nil {
				t.Fatalf("run: %v", err)
			}
		})
		expected, err := configformat.FormatAsEnv(cfg, root.ConfigFile, root.dbReg, configformat.AsOptions{})
		if err != nil {
			t.Fatalf("format: %v", err)
		}
		if output != expected {
			t.Fatalf("output mismatch\nexpected: %q\ngot: %q", expected, output)
		}
	})

	t.Run("as-json", func(t *testing.T) {
		cmd, err := parseConfigAsCmd(parent, "as-json", []string{})
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		output := captureStdout(t, func() {
			if err := cmd.asJSON(); err != nil {
				t.Fatalf("run: %v", err)
			}
		})
		expected, err := configformat.FormatAsJSON(cfg, root.ConfigFile)
		if err != nil {
			t.Fatalf("format: %v", err)
		}
		if output != expected {
			t.Fatalf("output mismatch\nexpected: %q\ngot: %q", expected, output)
		}
	})
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, reader)
		outCh <- buf.String()
	}()

	fn()
	_ = writer.Close()
	os.Stdout = old
	output := <-outCh
	_ = reader.Close()
	return output
}
