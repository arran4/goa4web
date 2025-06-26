package runtimeconfig

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestPaginationConfigPrecedence(t *testing.T) {
	env := map[string]string{
		config.EnvPageSizeMin:     "5",
		config.EnvPageSizeMax:     "30",
		config.EnvPageSizeDefault: "25",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Int("page-size-min", 12, "")
	fs.Int("page-size-default", 15, "")
	vals := map[string]string{
		config.EnvPageSizeMin:     "8",
		config.EnvPageSizeMax:     "20",
		config.EnvPageSizeDefault: "18",
	}
	_ = fs.Parse([]string{"--page-size-min=12", "--page-size-default=15"})
	cfg := GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })
	if cfg.PageSizeMin != 12 || cfg.PageSizeMax != 20 || cfg.PageSizeDefault != 15 {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadPaginationConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvPageSizeMin:     "7",
		config.EnvPageSizeDefault: "9",
	}
	cfg := GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.PageSizeMin != 7 || cfg.PageSizeDefault != 9 {
		t.Fatalf("want 7/9 got %#v", cfg)
	}
}
