package main

import (
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestHTTPConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvListen, ":1")
	os.Setenv(config.EnvHostname, "http://env")
	defer os.Unsetenv(config.EnvListen)
	defer os.Unsetenv(config.EnvHostname)

	cliRuntimeConfig = RuntimeConfig{HTTPListen: ":3", HTTPHostname: "http://cli"}
	vals := map[string]string{
		config.EnvListen:   ":2",
		config.EnvHostname: "http://file",
	}
	cfg := loadRuntimeConfig(vals)

	if cfg.HTTPListen != ":3" || cfg.HTTPHostname != "http://cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadHTTPConfigFromFileValues(t *testing.T) {
	cliRuntimeConfig = RuntimeConfig{}
	vals := map[string]string{
		config.EnvListen: ":9",
	}
	cfg := loadRuntimeConfig(vals)
	if cfg.HTTPListen != ":9" {
		t.Fatalf("want :9 got %q", cfg.HTTPListen)
	}
}
