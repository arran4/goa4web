package goa4web

import (
	"testing"

	"github.com/arran4/goa4web/core"
)

func TestLoadAppConfigFile(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "app.conf"
	content := "DB_USER=dbuser\nEMAIL_PROVIDER=smtp\n"
	if err := fs.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	m := LoadAppConfigFile(fs, file)
	if m["DB_USER"] != "dbuser" || m["EMAIL_PROVIDER"] != "smtp" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestLoadAppConfigFileMissing(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "none.conf"
	m := LoadAppConfigFile(fs, file)
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %#v", m)
	}
}
