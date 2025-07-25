package config_test

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestDefaultLanguageConfigPrecedence(t *testing.T) {
	env := map[string]string{config.EnvDefaultLanguage: "env"}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("default-language", "cli", "")
	vals := map[string]string{config.EnvDefaultLanguage: "file"}
	_ = fs.Parse([]string{"--default-language=cli"})

	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)
	if cfg.DefaultLanguage != "cli" {
		t.Fatalf("merged %#v", cfg.DefaultLanguage)
	}
}

func TestLoadDefaultLanguageFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{config.EnvDefaultLanguage: "fileval"}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.DefaultLanguage != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DefaultLanguage)
	}
}
