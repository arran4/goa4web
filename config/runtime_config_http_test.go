package config_test

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestHTTPConfigPrecedence(t *testing.T) {
	env := map[string]string{
		config.EnvListen:   ":1",
		config.EnvHostname: "http://env",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("listen", ":3", "")
	fs.String("hostname", "http://cli", "")
	vals := map[string]string{
		config.EnvListen:   ":2",
		config.EnvHostname: "http://file",
	}
	_ = fs.Parse([]string{"--listen=:3", "--hostname=http://cli"})
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)

	if cfg.HTTPListen != ":3" || cfg.HTTPHostname != "http://cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadHTTPConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvListen: ":9",
	}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.HTTPListen != ":9" {
		t.Fatalf("want :9 got %q", cfg.HTTPListen)
	}
}

func TestHTTPHostnameWithTrailingSlash(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvHostname: "http://example.com/",
	}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.HTTPHostname != "http://example.com" {
		t.Fatalf("want http://example.com got %q", cfg.HTTPHostname)
	}
}
