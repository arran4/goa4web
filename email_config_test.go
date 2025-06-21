package main

import (
	"testing"
)

func TestLoadEmailConfigFile(t *testing.T) {
	useMemFS(t)
	file := "email.conf"
	content := "EMAIL_PROVIDER=smtp\nSMTP_HOST=host\nSMTP_PORT=2525\n"
	if err := writeFile(file, []byte(content), 0644); err != nil {
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

func TestLoadEmailConfigEnvPath(t *testing.T) {
	useMemFS(t)
	file := "email.conf"
	if err := writeFile(file, []byte("EMAIL_PROVIDER=log\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	t.Setenv("EMAIL_CONFIG_FILE", file)
	emailConfigFile = ""
	cliEmailConfig = EmailConfig{}
	cfg := loadEmailConfig()
	if cfg.Provider != "log" {
		t.Fatalf("want log got %q", cfg.Provider)
	}
}
