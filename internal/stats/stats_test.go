package stats

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestStats(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	Inc("test_counter")
	Add("test_counter", 2)
	Inc("another_counter")

	Dump()

	output := buf.String()
	if !strings.Contains(output, "Stats dump:") {
		t.Errorf("Expected stats dump header, got: %s", output)
	}
	if !strings.Contains(output, "test_counter: 3") {
		t.Errorf("Expected test_counter: 3, got: %s", output)
	}
	if !strings.Contains(output, "another_counter: 1") {
		t.Errorf("Expected another_counter: 1, got: %s", output)
	}

	// Verify reset
	buf.Reset()
	Dump()
	output = buf.String()
	if output != "" {
		t.Errorf("Expected empty dump after reset, got: %s", output)
	}
}
