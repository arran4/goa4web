package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

func TestLoadAppConfigFile(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "app.env"
	content := "DB_USER=dbuser\nEMAIL_PROVIDER=smtp\n"
	if err := fs.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	m, err := config.LoadAppConfigFile(fs, file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if m["DB_USER"] != "dbuser" || m["EMAIL_PROVIDER"] != "smtp" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestLoadAppConfigFileMissing(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "none.env"
	m, err := config.LoadAppConfigFile(fs, file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %#v", m)
	}
}

func TestLoadAppConfigFileJSON(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "cfg.json"
	content := `{"DB_USER":"u","DB_PASS":"p"}`
	if err := fs.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	m, err := config.LoadAppConfigFile(fs, file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if m["DB_USER"] != "u" || m["DB_PASS"] != "p" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestLoadAppConfigFileUnsupported(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "cfg.txt"
	if err := fs.WriteFile(file, []byte("foo=bar"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := config.LoadAppConfigFile(fs, file); err == nil {
		t.Fatal("expected error")
	}
}
