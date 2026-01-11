package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestGenerateRuntimeConfigWithInjectedOptions(t *testing.T) {
	env := map[string]string{
		config.EnvDBConn:            "env",
		config.EnvDBLogVerbosity:    "2",
		config.EnvEmailLogVerbosity: "1",
		config.EnvFeedsEnabled:      "1",
	}

	strOpt := config.StringOption{Name: "db-conn-alt", Env: config.EnvDBConn, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *string { return &c.DBConn }}
	intOpt := config.IntOption{Name: "db-verb-alt", Env: config.EnvDBLogVerbosity, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *int { return &c.DBLogVerbosity }}
	intOpt2 := config.IntOption{Name: "email-verb-alt", Env: config.EnvEmailLogVerbosity, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *int { return &c.EmailLogVerbosity }}
	boolOpt := config.BoolOption{Name: "feeds-alt", Env: config.EnvFeedsEnabled, Usage: "", ExtendedUsage: "", Default: true, Target: func(c *config.RuntimeConfig) *bool { return &c.FeedsEnabled }}

	fs := config.NewRuntimeFlagSet("test")
	fs.String(strOpt.Name, strOpt.Default, strOpt.Usage)
	fs.Int(intOpt.Name, intOpt.Default, intOpt.Usage)
	fs.Int(intOpt2.Name, intOpt2.Default, intOpt2.Usage)
	fs.String(boolOpt.Name, "", boolOpt.Usage)
	_ = fs.Parse([]string{"--db-conn-alt=cli", "--db-verb-alt=5", "--email-verb-alt=4", "--feeds-alt=0"})

	vals := map[string]string{
		config.EnvDBConn:            "file",
		config.EnvDBLogVerbosity:    "3",
		config.EnvEmailLogVerbosity: "2",
		config.EnvFeedsEnabled:      "0",
	}

	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
		config.WithStringOptions([]config.StringOption{strOpt}),
		config.WithIntOptions([]config.IntOption{intOpt, intOpt2}),
		config.WithBoolOptions([]config.BoolOption{boolOpt}),
	)

	if cfg.DBConn != "cli" || cfg.DBLogVerbosity != 5 || cfg.EmailLogVerbosity != 4 || cfg.FeedsEnabled {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestGenerateRuntimeConfigWithInjectedFileValue(t *testing.T) {
	strOpt := config.StringOption{Name: "db-conn-alt", Env: config.EnvDBConn, Usage: "", ExtendedUsage: "", Target: func(c *config.RuntimeConfig) *string { return &c.DBConn }}
	boolOpt := config.BoolOption{Name: "feeds-alt", Env: config.EnvFeedsEnabled, Usage: "", ExtendedUsage: "", Default: true, Target: func(c *config.RuntimeConfig) *bool { return &c.FeedsEnabled }}
	fs := config.NewRuntimeFlagSet("test")
	fs.String(strOpt.Name, strOpt.Default, strOpt.Usage)
	fs.String(boolOpt.Name, "", boolOpt.Usage)
	_ = fs.Parse(nil)

	vals := map[string]string{config.EnvDBConn: "file", config.EnvFeedsEnabled: "0"}

	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(string) string { return "" }),
		config.WithStringOptions([]config.StringOption{strOpt}),
		config.WithBoolOptions([]config.BoolOption{boolOpt}),
	)

	if cfg.DBConn != "file" || cfg.FeedsEnabled {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestGenerateRuntimeConfigMigrationsDirPrecedence(t *testing.T) {
	t.Run("flag-overrides-file-and-env", func(t *testing.T) {
		fs := config.NewRuntimeFlagSet("test")
		_ = fs.Parse([]string{"--migrations-dir=cli"})

		cfg := config.NewRuntimeConfig(
			config.WithFlagSet(fs),
			config.WithFileValues(map[string]string{config.EnvMigrationsDir: "file"}),
			config.WithGetenv(func(k string) string {
				if k == config.EnvMigrationsDir {
					return "env"
				}
				return ""
			}),
		)

		if cfg.MigrationsDir != "cli" {
			t.Fatalf("expected flag to win, got %q", cfg.MigrationsDir)
		}
	})

	t.Run("file-overrides-env", func(t *testing.T) {
		fs := config.NewRuntimeFlagSet("test")
		_ = fs.Parse(nil)

		cfg := config.NewRuntimeConfig(
			config.WithFlagSet(fs),
			config.WithFileValues(map[string]string{config.EnvMigrationsDir: "file"}),
			config.WithGetenv(func(k string) string {
				if k == config.EnvMigrationsDir {
					return "env"
				}
				return ""
			}),
		)

		if cfg.MigrationsDir != "file" {
			t.Fatalf("expected file to win, got %q", cfg.MigrationsDir)
		}
	})

	t.Run("env-used-when-flag-and-file-missing", func(t *testing.T) {
		fs := config.NewRuntimeFlagSet("test")
		_ = fs.Parse(nil)

		cfg := config.NewRuntimeConfig(
			config.WithFlagSet(fs),
			config.WithFileValues(map[string]string{}),
			config.WithGetenv(func(k string) string {
				if k == config.EnvMigrationsDir {
					return "env"
				}
				return ""
			}),
		)

		if cfg.MigrationsDir != "env" {
			t.Fatalf("expected env to win, got %q", cfg.MigrationsDir)
		}
	})
}
