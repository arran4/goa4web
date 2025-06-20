package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmailConfigFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "email.conf")
	content := "EMAIL_PROVIDER=smtp\nSMTP_HOST=host\nSMTP_PORT=2525\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loadEmailConfigFile(file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Provider != "smtp" || cfg.SMTPHost != "host" || cfg.SMTPPort != "2525" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestResolveEmailConfigPrecedence(t *testing.T) {
	env := EmailConfig{Provider: "ses", SMTPHost: "env"}
	file := EmailConfig{Provider: "log", SMTPHost: "file"}
	cli := EmailConfig{Provider: "smtp", SMTPPort: "25"}

	cfg := resolveEmailConfig(cli, file, env)

	if cfg.Provider != "smtp" || cfg.SMTPHost != "file" || cfg.SMTPPort != "25" {
		t.Fatalf("merged %#v", cfg)
	}
}
