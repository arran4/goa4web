package main

import (
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestDBConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvDBUser, "env")
	os.Setenv(config.EnvDBHost, "env")
	defer os.Unsetenv(config.EnvDBUser)
	defer os.Unsetenv(config.EnvDBHost)

	cliRuntimeConfig = RuntimeConfig{DBPass: "cli"}
	vals := map[string]string{
		config.EnvDBUser: "file",
		config.EnvDBPort: "1",
	}
	cfg := loadRuntimeConfig(vals)
	if cfg.DBUser != "file" || cfg.DBPass != "cli" || cfg.DBHost != "env" || cfg.DBPort != "1" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadDBConfigFromFileValues(t *testing.T) {
	cliRuntimeConfig = RuntimeConfig{}
	vals := map[string]string{
		config.EnvDBUser: "fileval",
	}
	cfg := loadRuntimeConfig(vals)
	if cfg.DBUser != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DBUser)
	}
}
