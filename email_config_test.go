package main

import (
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestEmailConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvEmailProvider, "ses")
	os.Setenv(config.EnvSMTPHost, "env")
	defer os.Unsetenv(config.EnvEmailProvider)
	defer os.Unsetenv(config.EnvSMTPHost)

	cliRuntimeConfig = RuntimeConfig{
		EmailProvider: "smtp",
		EmailSMTPPort: "25",
	}
	vals := map[string]string{
		config.EnvEmailProvider: "log",
		config.EnvSMTPHost:      "file",
	}
	cfg := loadRuntimeConfig(vals)
	if cfg.EmailProvider != "smtp" || cfg.EmailSMTPHost != "file" || cfg.EmailSMTPPort != "25" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadEmailConfigFromFileValues(t *testing.T) {
	cliRuntimeConfig = RuntimeConfig{}
	vals := map[string]string{
		config.EnvEmailProvider: "log",
	}
	cfg := loadRuntimeConfig(vals)
	if cfg.EmailProvider != "log" {
		t.Fatalf("want log got %q", cfg.EmailProvider)
	}
}
