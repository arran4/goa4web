package config_test

import (
	"encoding/json"
	"github.com/arran4/goa4web/config"
	"testing"

	"github.com/arran4/goa4web/core"
)

func TestAddMissingJSONOptions(t *testing.T) {
	fs := core.UseMemFS(t)
	path := "cfg.json"
	initial := `{"DB_USER":"user"}`
	if err := fs.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	err := config.AddMissingJSONOptions(fs, path, map[string]string{
		"DB_USER": "user",
		"DB_PASS": "pass",
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	b, err := fs.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["DB_USER"] != "user" || m["DB_PASS"] != "pass" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestAddMissingJSONOptionsCreate(t *testing.T) {
	fs := core.UseMemFS(t)
	path := "new.json"
	err := config.AddMissingJSONOptions(fs, path, map[string]string{"DB_USER": "u"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	b, err := fs.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["DB_USER"] != "u" {
		t.Fatalf("unexpected map: %#v", m)
	}
}
