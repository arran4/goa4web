package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleDie(t *testing.T) {
	rr := httptest.NewRecorder()
	handleDie(rr, "oops")
	if ct := rr.Header().Get("Content-Type"); ct != "text/html" {
		t.Errorf("expected Content-Type text/html, got %q", ct)
	}
	expected := "<b><font color=red>You encountered an error: oops....</font></b>"
	if rr.Body.String() != expected {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestConfigurationSetGet(t *testing.T) {
	c := NewConfiguration()
	c.set("foo", "bar")
	if got := c.get("foo"); got != "bar" {
		t.Errorf("get(foo)=%q want bar", got)
	}
}

func TestConfigurationRead(t *testing.T) {
	dir := t.TempDir()
	fname := filepath.Join(dir, "conf.txt")
	content := "k1=v1\nk2=v=2\ninvalid\n spaced = value with spaces\n"
	if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	c := NewConfiguration()
	c.readConfiguration(fname)
	if got := c.get("k1"); got != "v1" {
		t.Errorf("k1=%q want v1", got)
	}
	if got := c.get("k2"); got != "v=2" {
		t.Errorf("k2=%q want v=2", got)
	}
	if got := c.get("invalid"); got != "" {
		t.Errorf("invalid=%q want empty", got)
	}
	if got := c.get(" spaced "); got != " value with spaces" {
		t.Errorf("spaced=%q", got)
	}
}

func TestX2c(t *testing.T) {
	if X2c("41") != 'A' {
		t.Errorf("expected 0x41")
	}
	if X2c("0a") != 0x0a {
		t.Errorf("expected 0x0a")
	}
	if X2c("G1") != 1 {
		t.Errorf("expected 1 for invalid hex")
	}
}
