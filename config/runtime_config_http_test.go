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

	if cfg.HTTPListen != ":3" || cfg.BaseURL != "http://cli" {
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
	if cfg.BaseURL != "http://example.com" {
		t.Fatalf("want http://example.com got %q", cfg.BaseURL)
	}
}

func TestBaseURLPrecedence(t *testing.T) {
	// 1. ExternalURL takes precedence
	cfg := config.NewRuntimeConfig(
		config.WithFileValues(map[string]string{
			config.EnvExternalURL: "http://external",
			config.EnvHostname:    "http://hostname",
			config.EnvHost:        "host",
		}),
	)
	if cfg.BaseURL != "http://external" {
		t.Errorf("expected ExternalURL precedence, got %q", cfg.BaseURL)
	}

	// 2. HTTPHostname takes precedence over Host
	cfg = config.NewRuntimeConfig(
		config.WithFileValues(map[string]string{
			config.EnvHostname: "http://hostname",
			config.EnvHost:     "host",
		}),
	)
	if cfg.BaseURL != "http://hostname" {
		t.Errorf("expected HTTPHostname precedence, got %q", cfg.BaseURL)
	}

	// 3. Host is used if others missing
	cfg = config.NewRuntimeConfig(
		config.WithFileValues(map[string]string{
			config.EnvHost: "host",
		}),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.BaseURL != "http://host" {
		t.Errorf("expected Host usage, got %q", cfg.BaseURL)
	}
}

func TestHostPreferredWhenHostnameLacksScheme(t *testing.T) {
	cfg := config.NewRuntimeConfig(
		config.WithFileValues(map[string]string{
			config.EnvHost: "configured-host",
		}),
		config.WithGetenv(func(key string) string {
			if key == config.EnvHostname {
				return "container-hostname"
			}
			return ""
		}),
	)

	if cfg.BaseURL != "http://configured-host" {
		t.Fatalf("expected host to win over non-URL hostname, got %q", cfg.BaseURL)
	}
}
