package goa4web

import (
	"flag"
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestPaginationConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvPageSizeMin, "5")
	os.Setenv(config.EnvPageSizeMax, "30")
	os.Setenv(config.EnvPageSizeDefault, "25")
	defer os.Unsetenv(config.EnvPageSizeMin)
	defer os.Unsetenv(config.EnvPageSizeMax)
	defer os.Unsetenv(config.EnvPageSizeDefault)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Int("page-size-min", 12, "")
	fs.Int("page-size-default", 15, "")
	vals := map[string]string{
		config.EnvPageSizeMin:     "8",
		config.EnvPageSizeMax:     "20",
		config.EnvPageSizeDefault: "18",
	}
	_ = fs.Parse([]string{"--page-size-min=12", "--page-size-default=15"})
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals)
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
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals)
	if cfg.PageSizeMin != 7 || cfg.PageSizeDefault != 9 {
		t.Fatalf("want 7/9 got %#v", cfg)
	}
}
