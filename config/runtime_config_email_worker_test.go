package config_test

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestEmailWorkerIntervalPrecedence(t *testing.T) {
	env := map[string]string{config.EnvEmailWorkerInterval: "30"}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Int("email-worker-interval", 15, "")
	vals := map[string]string{config.EnvEmailWorkerInterval: "20"}
	_ = fs.Parse([]string{"--email-worker-interval=15"})

	cfg := config.GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })
	if cfg.EmailWorkerInterval != 15 {
		t.Fatalf("merged %#v", cfg.EmailWorkerInterval)
	}
}

func TestLoadEmailWorkerIntervalFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{config.EnvEmailWorkerInterval: "25"}

	cfg := config.GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.EmailWorkerInterval != 25 {
		t.Fatalf("want 25 got %d", cfg.EmailWorkerInterval)
	}
}
