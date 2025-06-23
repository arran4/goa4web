package main

import (
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestLanguageConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvDefaultLanguage, "env")
	defer os.Unsetenv(config.EnvDefaultLanguage)

	cliRuntimeConfig = RuntimeConfig{DefaultLanguage: "cli"}
	vals := map[string]string{config.EnvDefaultLanguage: "file"}
	cfg := loadRuntimeConfig(vals)
	if cfg.DefaultLanguage != "cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadLanguageConfigFromFileValues(t *testing.T) {
	cliRuntimeConfig = RuntimeConfig{}
	vals := map[string]string{config.EnvDefaultLanguage: "file"}
	cfg := loadRuntimeConfig(vals)
	if cfg.DefaultLanguage != "file" {
		t.Fatalf("want file got %q", cfg.DefaultLanguage)
	}
}
