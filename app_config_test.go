package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppConfigFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.conf")
	content := "DB_CONFIG_FILE=db.conf\nEMAIL_CONFIG_FILE=email.conf\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	m := loadAppConfigFile(file)
	if m["DB_CONFIG_FILE"] != "db.conf" || m["EMAIL_CONFIG_FILE"] != "email.conf" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestLoadAppConfigFileMissing(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "none.conf")
	m := loadAppConfigFile(file)
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %#v", m)
	}
}
