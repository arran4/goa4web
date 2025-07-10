package config

import (
	"flag"
	"testing"
)

func TestDefaultLanguageConfigPrecedence(t *testing.T) {
	env := map[string]string{EnvDefaultLanguage: "env"}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("default-language", "cli", "")
	vals := map[string]string{EnvDefaultLanguage: "file"}
	_ = fs.Parse([]string{"--default-language=cli"})

	cfg := GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })
	if cfg.DefaultLanguage != "cli" {
		t.Fatalf("merged %#v", cfg.DefaultLanguage)
	}
}

func TestLoadDefaultLanguageFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{EnvDefaultLanguage: "fileval"}
	cfg := GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.DefaultLanguage != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DefaultLanguage)
	}
}
