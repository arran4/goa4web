package goa4web

import (
	"flag"
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestDefaultLanguageConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvDefaultLanguage, "env")
	defer os.Unsetenv(config.EnvDefaultLanguage)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("default-language", "cli", "")
	vals := map[string]string{config.EnvDefaultLanguage: "file"}
	_ = fs.Parse([]string{"--default-language=cli"})

	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals)
	if cfg.DefaultLanguage != "cli" {
		t.Fatalf("merged %#v", cfg.DefaultLanguage)
	}
}

func TestLoadDefaultLanguageFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{config.EnvDefaultLanguage: "fileval"}
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals)
	if cfg.DefaultLanguage != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DefaultLanguage)
	}
}
