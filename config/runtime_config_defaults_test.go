package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestRuntimeConfigDefaultsFromOptions(t *testing.T) {
	cfg := config.NewRuntimeConfig(
		config.WithGetenv(func(string) string { return "" }),
	)

	if cfg.HTTPListen != ":8080" {
		t.Fatalf("http listen default = %q", cfg.HTTPListen)
	}
	if cfg.DBTimezone != "Australia/Melbourne" {
		t.Fatalf("db timezone default = %q", cfg.DBTimezone)
	}
	if cfg.SessionName != "my-session" {
		t.Fatalf("session name default = %q", cfg.SessionName)
	}
	if cfg.SessionSameSite != "strict" {
		t.Fatalf("session same site default = %q", cfg.SessionSameSite)
	}
	if cfg.PageSizeMin != 5 || cfg.PageSizeMax != 50 || cfg.PageSizeDefault != config.DefaultPageSize {
		t.Fatalf("page size defaults = %d/%d/%d", cfg.PageSizeMin, cfg.PageSizeMax, cfg.PageSizeDefault)
	}
	if cfg.StatsStartYear != 2005 {
		t.Fatalf("stats start year default = %d", cfg.StatsStartYear)
	}
	if cfg.Timezone != "Australia/Melbourne" {
		t.Fatalf("timezone default = %q", cfg.Timezone)
	}
	if cfg.HSTSHeaderValue != "max-age=63072000; includeSubDomains" {
		t.Fatalf("hsts header default = %q", cfg.HSTSHeaderValue)
	}
	if cfg.ImageUploadProvider != "local" {
		t.Fatalf("image upload provider default = %q", cfg.ImageUploadProvider)
	}
	if cfg.ImageCacheProvider != "local" {
		t.Fatalf("image cache provider default = %q", cfg.ImageCacheProvider)
	}
	if cfg.LoginAttemptWindow != 15 || cfg.LoginAttemptThreshold != 5 {
		t.Fatalf("login attempt defaults = %d/%d", cfg.LoginAttemptWindow, cfg.LoginAttemptThreshold)
	}
}
