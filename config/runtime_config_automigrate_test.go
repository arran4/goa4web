package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestAutoMigratePrecedence(t *testing.T) {
	env := map[string]string{config.EnvAutoMigrate: "0"}
	vals := map[string]string{config.EnvAutoMigrate: "1"}

	fs := config.NewRuntimeFlagSet("test")
	_ = fs.Parse(nil)
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)
	if !cfg.AutoMigrate {
		t.Fatalf("expected file value to override env")
	}

	fs = config.NewRuntimeFlagSet("test")
	_ = fs.Parse([]string{"--auto-migrate=0"})
	cfg = config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)
	if cfg.AutoMigrate {
		t.Fatalf("expected CLI value to override file and env")
	}
}
