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
	if cfg.ImageCachePlaceholderMinWidth != config.DefaultImageCachePlaceholderMinWidth || cfg.ImageCachePlaceholderMinHeight != config.DefaultImageCachePlaceholderMinHeight {
		t.Fatalf("image cache placeholder defaults = %dx%d", cfg.ImageCachePlaceholderMinWidth, cfg.ImageCachePlaceholderMinHeight)
	}
	if sizes := cfg.ThumbnailSizes(); len(sizes) != 2 || sizes[0] != (config.ThumbnailSize{Width: config.DefaultImageThumbnailWidth, Height: config.DefaultImageThumbnailHeight}) || sizes[1] != (config.ThumbnailSize{Width: 2048, Height: 1600}) {
		t.Fatalf("thumbnail sizes = %v", sizes)
	}
	if cfg.LoginAttemptWindow != 15 || cfg.LoginAttemptThreshold != 5 {
		t.Fatalf("login attempt defaults = %d/%d", cfg.LoginAttemptWindow, cfg.LoginAttemptThreshold)
	}
	if cfg.ForumPostAppendWindow != 60 || cfg.PrivateForumPostAppendWindow != 60 {
		t.Fatalf("forum append window defaults = %d/%d", cfg.ForumPostAppendWindow, cfg.PrivateForumPostAppendWindow)
	}
}

func TestForumAppendWindowConfigurationPrecedenceAndIndependence(t *testing.T) {
	fs := config.NewRuntimeFlagSet("test")
	if err := fs.Parse([]string{"--forum-post-append-window=15"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(map[string]string{
			config.EnvForumPostAppendWindow:        "30",
			config.EnvPrivateForumPostAppendWindow: "45",
		}),
		config.WithGetenv(func(key string) string {
			switch key {
			case config.EnvForumPostAppendWindow:
				return "50"
			case config.EnvPrivateForumPostAppendWindow:
				return "55"
			default:
				return ""
			}
		}),
	)
	if cfg.ForumPostAppendWindow != 15 || cfg.PrivateForumPostAppendWindow != 45 {
		t.Fatalf("append windows = %d/%d, want flag 15 and file 45", cfg.ForumPostAppendWindow, cfg.PrivateForumPostAppendWindow)
	}

	envOnly := config.NewRuntimeConfig(
		config.WithGetenv(func(key string) string {
			if key == config.EnvPrivateForumPostAppendWindow {
				return "0"
			}
			return ""
		}),
	)
	if envOnly.ForumPostAppendWindow != 60 || envOnly.PrivateForumPostAppendWindow != 0 {
		t.Fatalf("independent env append windows = %d/%d", envOnly.ForumPostAppendWindow, envOnly.PrivateForumPostAppendWindow)
	}
}

func TestThumbnailSizes(t *testing.T) {
	cfg := &config.RuntimeConfig{ImageThumbnailSize: 100, ImageThumbnailSizes: "128x64, 256x128, 128x64, 100, invalid, 0"}
	got := cfg.ThumbnailSizes()
	want := []config.ThumbnailSize{{Width: 128, Height: 64}, {Width: 256, Height: 128}, {Width: 100, Height: 100}}
	if len(got) != len(want) {
		t.Fatalf("thumbnail sizes = %v, want %v", got, want)
	}
	for i, size := range want {
		if got[i] != size {
			t.Fatalf("thumbnail sizes = %v, want %v", got, want)
		}
	}
}
