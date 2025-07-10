package config_test

import (
	"github.com/arran4/goa4web/config"
	"testing"
)

func TestParseEnvBytes(t *testing.T) {
	data := []byte("DB_USER=user # comment\n# full comment\n\nDB_PASS=pass\n")
	vals := config.ParseEnvBytes(data)
	if vals["DB_USER"] != "user" || vals["DB_PASS"] != "pass" {
		t.Fatalf("unexpected values: %#v", vals)
	}
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}
}
