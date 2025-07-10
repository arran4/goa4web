package config

import (
	"flag"
	"testing"
)

func TestDBConfigPrecedence(t *testing.T) {
	env := map[string]string{
		EnvDBConn: "env",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("db-conn", "cli", "")
	vals := map[string]string{
		EnvDBConn: "file",
	}
	_ = fs.Parse([]string{"--db-conn=cli"})
	cfg := GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })
	if cfg.DBConn != "cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadDBConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		EnvDBConn: "fileval",
	}
	cfg := GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.DBConn != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DBConn)
	}
}
