package main

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestLoadEmailConfigFile(t *testing.T) {
	fsys := fstest.MapFS{
		"email.conf": {Data: []byte("EMAIL_PROVIDER=smtp\nSMTP_HOST=host\nSMTP_PORT=2525\n")},
	}
	old := emailReadFile
	emailReadFile = func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }
	defer func() { emailReadFile = old }()
	cfg, err := loadEmailConfigFile("email.conf")
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
	fsys := fstest.MapFS{
		"email.conf": {Data: []byte("EMAIL_PROVIDER=log\n")},
	}
	old := emailReadFile
	emailReadFile = func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }
	defer func() { emailReadFile = old }()
	t.Setenv("EMAIL_CONFIG_FILE", "email.conf")
	emailConfigFile = ""
	cliEmailConfig = EmailConfig{}
	cfg := loadEmailConfig()
	if cfg.Provider != "log" {
		t.Fatalf("want log got %q", cfg.Provider)
	}
}
