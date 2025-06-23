package main

import (
	"flag"
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestEmailConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvEmailProvider, "ses")
	os.Setenv(config.EnvSMTPHost, "env")
	defer os.Unsetenv(config.EnvEmailProvider)
	defer os.Unsetenv(config.EnvSMTPHost)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("email-provider", "smtp", "")
	fs.String("smtp-port", "25", "")
	vals := map[string]string{
		config.EnvEmailProvider: "log",
		config.EnvSMTPHost:      "file",
	}
	_ = fs.Parse([]string{"--email-provider=smtp", "--smtp-port=25"})
	cfg := generateRuntimeConfig(fs, vals)
	if cfg.EmailProvider != "smtp" || cfg.EmailSMTPHost != "file" || cfg.EmailSMTPPort != "25" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadEmailConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvEmailProvider: "log",
	}
	cfg := generateRuntimeConfig(fs, vals)
	if cfg.EmailProvider != "log" {
		t.Fatalf("want log got %q", cfg.EmailProvider)
	}
}
