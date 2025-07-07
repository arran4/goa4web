package runtimeconfig

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestGenerateRuntimeConfigWithInjectedOptions(t *testing.T) {
	env := map[string]string{
		config.EnvDBConn:         "env",
		config.EnvDBLogVerbosity: "2",
	}

	strOpt := StringOption{Name: "db-conn-alt", Env: config.EnvDBConn, Field: "DBConn", Usage: ""}
	intOpt := IntOption{Name: "db-verb-alt", Env: config.EnvDBLogVerbosity, Field: "DBLogVerbosity", Usage: ""}
	fs := NewRuntimeFlagSetWithOptions("test", []StringOption{strOpt}, []IntOption{intOpt})
	_ = fs.Parse([]string{"--db-conn-alt=cli", "--db-verb-alt=5"})

	vals := map[string]string{
		config.EnvDBConn:         "file",
		config.EnvDBLogVerbosity: "3",
	}

	cfg := GenerateRuntimeConfigWithOptions(fs, vals, func(k string) string { return env[k] }, []StringOption{strOpt}, []IntOption{intOpt})

	if cfg.DBConn != "cli" || cfg.DBLogVerbosity != 5 {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestGenerateRuntimeConfigWithInjectedFileValue(t *testing.T) {
	strOpt := StringOption{Name: "db-conn-alt", Env: config.EnvDBConn, Field: "DBConn", Usage: ""}
	fs := NewRuntimeFlagSetWithOptions("test", []StringOption{strOpt}, nil)
	_ = fs.Parse(nil)

	vals := map[string]string{config.EnvDBConn: "file"}

	cfg := GenerateRuntimeConfigWithOptions(fs, vals, func(string) string { return "" }, []StringOption{strOpt}, nil)

	if cfg.DBConn != "file" {
		t.Fatalf("want file got %q", cfg.DBConn)
	}
}
