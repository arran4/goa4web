package main

import (
	"flag"
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestHTTPConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvListen, ":1")
	os.Setenv(config.EnvHostname, "http://env")
	defer os.Unsetenv(config.EnvListen)
	defer os.Unsetenv(config.EnvHostname)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("listen", ":3", "")
	fs.String("hostname", "http://cli", "")
	vals := map[string]string{
		config.EnvListen:   ":2",
		config.EnvHostname: "http://file",
	}
	_ = fs.Parse([]string{"--listen=:3", "--hostname=http://cli"})
	cfg := generateRuntimeConfig(fs, vals)

	if cfg.HTTPListen != ":3" || cfg.HTTPHostname != "http://cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadHTTPConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvListen: ":9",
	}
	cfg := generateRuntimeConfig(fs, vals)
	if cfg.HTTPListen != ":9" {
		t.Fatalf("want :9 got %q", cfg.HTTPListen)
	}
}
