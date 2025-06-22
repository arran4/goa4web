package main

import (
	"testing"
)

func TestLoadAppConfigFile(t *testing.T) {
	useMemFS(t)
	file := "app.conf"
	content := "DB_CONFIG_FILE=db.conf\nEMAIL_CONFIG_FILE=email.conf\n"
	if err := writeFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	m := loadAppConfigFile(file)
	if m["DB_CONFIG_FILE"] != "db.conf" || m["EMAIL_CONFIG_FILE"] != "email.conf" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestLoadAppConfigFileMissing(t *testing.T) {
	useMemFS(t)
	file := "none.conf"
	m := loadAppConfigFile(file)
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %#v", m)
	}
}
