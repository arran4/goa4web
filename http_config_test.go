package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadHTTPConfigFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "http.conf")
	if err := os.WriteFile(file, []byte("LISTEN=1.2.3.4:80\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loadHTTPConfigFile(file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Listen != "1.2.3.4:80" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestResolveHTTPConfigPrecedence(t *testing.T) {
	env := HTTPConfig{Listen: ":1"}
	file := HTTPConfig{Listen: ":2"}
	cli := HTTPConfig{Listen: ":3"}

	cfg := resolveHTTPConfig(cli, file, env)

	if cfg.Listen != ":3" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadHTTPConfigEnvPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "http.conf")
	if err := os.WriteFile(file, []byte("LISTEN=:9\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HTTP_CONFIG_FILE", file)
	httpConfigFile = ""
	cliHTTPConfig = HTTPConfig{}
	cfg := loadHTTPConfig()
	if cfg.Listen != ":9" {
		t.Fatalf("want :9 got %q", cfg.Listen)
	}
}
