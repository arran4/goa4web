package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestGenerateRuntimeConfigWithInjectedOptions(t *testing.T) {
	env := map[string]string{
		config.EnvDBConn:         "env",
		config.EnvDBLogVerbosity: "2",
	}

	strOpt := config.StringOption{Name: "db-conn-alt", Env: config.EnvDBConn, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *string { return &c.DBConn }}
	intOpt := config.IntOption{Name: "db-verb-alt", Env: config.EnvDBLogVerbosity, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *int { return &c.DBLogVerbosity }}
	fs := config.NewRuntimeFlagSetWithOptions("test", []config.StringOption{strOpt}, []config.IntOption{intOpt})
	_ = fs.Parse([]string{"--db-conn-alt=cli", "--db-verb-alt=5"})

	vals := map[string]string{
		config.EnvDBConn:         "file",
		config.EnvDBLogVerbosity: "3",
	}

	cfg := config.GenerateRuntimeConfigWithOptions(fs, vals, func(k string) string { return env[k] }, []config.StringOption{strOpt}, []config.IntOption{intOpt})

	if cfg.DBConn != "cli" || cfg.DBLogVerbosity != 5 {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestGenerateRuntimeConfigWithInjectedFileValue(t *testing.T) {
	strOpt := config.StringOption{Name: "db-conn-alt", Env: config.EnvDBConn, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *string { return &c.DBConn }}
	fs := config.NewRuntimeFlagSetWithOptions("test", []config.StringOption{strOpt}, nil)
	_ = fs.Parse(nil)

	vals := map[string]string{config.EnvDBConn: "file"}

	cfg := config.GenerateRuntimeConfigWithOptions(fs, vals, func(string) string { return "" }, []config.StringOption{strOpt}, nil)

	if cfg.DBConn != "file" {
		t.Fatalf("want file got %q", cfg.DBConn)
	}
}
