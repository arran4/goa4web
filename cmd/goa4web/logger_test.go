package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestRootCmd_Logger_Caller(t *testing.T) {
	// Save original output and flags
	origOutput := log.Writer()
	origFlags := log.Flags()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(log.Lshortfile)

	defer func() {
		// Restore original output and flags
		log.SetOutput(origOutput)
		if origOutput == nil {
			log.SetOutput(os.Stderr)
		}
		log.SetFlags(origFlags)
	}()

	r := &rootCmd{Verbosity: 1}

	t.Run("Infof", func(t *testing.T) {
		buf.Reset()
		r.Infof("test message info")
		output := buf.String()

		if !strings.Contains(output, "logger_test.go") {
			t.Logf("Output was: %s", output)
			t.Errorf("Expected output to contain 'logger_test.go'")
		}
		if strings.Contains(output, "main.go") {
			t.Logf("Output was: %s", output)
			t.Error("Output incorrectly contains 'main.go'")
		}
	})

	t.Run("Verbosef", func(t *testing.T) {
		buf.Reset()
		r.Verbosef("test message verbose")
		output := buf.String()

		if !strings.Contains(output, "logger_test.go") {
			t.Logf("Output was: %s", output)
			t.Errorf("Expected output to contain 'logger_test.go'")
		}
		if strings.Contains(output, "main.go") {
			t.Logf("Output was: %s", output)
			t.Error("Output incorrectly contains 'main.go'")
		}
	})
}
