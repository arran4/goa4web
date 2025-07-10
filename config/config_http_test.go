package config

import (
	"flag"
	"testing"
)

func TestHTTPConfigPrecedence(t *testing.T) {
	env := map[string]string{
		EnvListen:   ":1",
		EnvHostname: "http://env",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("listen", ":3", "")
	fs.String("hostname", "http://cli", "")
	vals := map[string]string{
		EnvListen:   ":2",
		EnvHostname: "http://file",
	}
	_ = fs.Parse([]string{"--listen=:3", "--hostname=http://cli"})
	cfg := GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })

	if cfg.HTTPListen != ":3" || cfg.HTTPHostname != "http://cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadHTTPConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		EnvListen: ":9",
	}
	cfg := GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.HTTPListen != ":9" {
		t.Fatalf("want :9 got %q", cfg.HTTPListen)
	}
}
