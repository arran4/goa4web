package goa4web

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestDBConfigPrecedence(t *testing.T) {
	env := map[string]string{
		config.EnvDBUser: "env",
		config.EnvDBHost: "env",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("db-pass", "cli", "")
	vals := map[string]string{
		config.EnvDBUser: "file",
		config.EnvDBPort: "1",
	}
	_ = fs.Parse([]string{"--db-pass=cli"})
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })
	if cfg.DBUser != "file" || cfg.DBPass != "cli" || cfg.DBHost != "env" || cfg.DBPort != "1" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadDBConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvDBUser: "fileval",
	}
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.DBUser != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DBUser)
	}
}
