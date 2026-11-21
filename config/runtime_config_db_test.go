package config_test

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestDBConfigPrecedence(t *testing.T) {
	env := map[string]string{
		config.EnvDBConn: "env",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("db-conn", "cli", "")
	vals := map[string]string{
		config.EnvDBConn: "file",
	}
	_ = fs.Parse([]string{"--db-conn=cli"})
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)
	if cfg.DBConn != "cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadDBConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvDBConn: "fileval",
	}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.DBConn != "fileval" {
		t.Fatalf("want fileval got %q", cfg.DBConn)
	}
}

func TestDBTimezonePrecedence(t *testing.T) {
	env := map[string]string{
		config.EnvDBTimezone: "env",
	}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("db-timezone", "cli", "")
	vals := map[string]string{
		config.EnvDBTimezone: "file",
	}
	_ = fs.Parse([]string{"--db-timezone=cli"})
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)
	if cfg.DBTimezone != "cli" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestDefaultDBTimezone(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(map[string]string{}),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.DBTimezone != "Australia/Melbourne" {
		t.Fatalf("want Australia/Melbourne got %q", cfg.DBTimezone)
	}
}

func TestDBMigrationConnPrecedence(t *testing.T) {
	env := map[string]string{config.EnvDBMigrationConn: "environment"}
	vals := map[string]string{config.EnvDBMigrationConn: "file"}
	fs := config.NewRuntimeFlagSet("test")
	if err := fs.Parse([]string{"--db-migration-conn=cli"}); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(key string) string { return env[key] }),
	)
	if cfg.DBMigrationConn != "cli" {
		t.Fatalf("expected CLI migration DSN to take precedence; got %q", cfg.DBMigrationConn)
	}
	cfg = config.NewRuntimeConfig(
		config.WithFileValues(vals),
		config.WithGetenv(func(key string) string { return env[key] }),
	)
	if cfg.DBMigrationConn != "file" {
		t.Fatalf("expected file migration DSN to take precedence; got %q", cfg.DBMigrationConn)
	}
	cfg = config.NewRuntimeConfig(
		config.WithGetenv(func(key string) string { return env[key] }),
	)
	if cfg.DBMigrationConn != "environment" {
		t.Fatalf("expected env migration DSN; got %q", cfg.DBMigrationConn)
	}
}
