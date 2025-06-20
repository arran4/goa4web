package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDBConfigFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "db.conf")
	content := "DB_USER=u\nDB_PASS=p\nDB_HOST=h\nDB_PORT=3307\nDB_NAME=n\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loadDBConfigFile(file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.User != "u" || cfg.Pass != "p" || cfg.Host != "h" || cfg.Port != "3307" || cfg.Name != "n" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestResolveDBConfigPrecedence(t *testing.T) {
	env := DBConfig{User: "env", Host: "env"}
	file := DBConfig{User: "file", Port: "1"}
	cli := DBConfig{Pass: "cli"}

	cfg := resolveDBConfig(cli, file, env)

	if cfg.User != "file" || cfg.Pass != "cli" || cfg.Host != "env" || cfg.Port != "1" {
		t.Fatalf("merged %#v", cfg)
	}
}
