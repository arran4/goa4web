package goa4web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core"
)

func TestHandleDie(t *testing.T) {
	rr := httptest.NewRecorder()
	handleDie(rr, "oops")
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type text/plain; charset=utf-8, got %q", ct)
	}
	if body := rr.Body.String(); body != "oops\n" {
		t.Errorf("unexpected body: %q", body)
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
	fs := core.UseMemFS(t)
	fname := "conf.txt"
	content := "k1=v1\nk2=v=2\ninvalid\n spaced = value with spaces\n"
	if err := fs.WriteFile(fname, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	c := NewConfiguration()
	c.readConfiguration(fs, fname)
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
