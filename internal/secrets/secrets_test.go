package secrets

import (
	"path/filepath"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	t.Run("dev version", func(t *testing.T) {
		got := DefaultPath("secret1", "MY_DOCKER_ENV", WithVersion("dev"))
		if got != ".secret1" {
			t.Errorf("expected %q, got %q", ".secret1", got)
		}
	})

	t.Run("docker env set", func(t *testing.T) {
		got := DefaultPath("secret1", "MY_DOCKER_ENV",
			WithVersion("v1.0.0"),
			WithGetenv(func(k string) string {
				if k == "MY_DOCKER_ENV" {
					return "1"
				}
				return ""
			}),
		)
		expected := filepath.Join("/var/lib/goa4web", "secret1")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("no home or xdg", func(t *testing.T) {
		got := DefaultPath("secret1", "MISSING_DOCKER_ENV",
			WithVersion("v1.0.0"),
			WithGetenv(func(k string) string {
				return ""
			}),
		)
		expected := filepath.Join("/var/lib/goa4web", "secret1")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("user config dir", func(t *testing.T) {
		got := DefaultPath("secret1", "MISSING_DOCKER_ENV",
			WithVersion("v1.0.0"),
			WithGetenv(func(k string) string {
				if k == "HOME" {
					return "/fake/home"
				}
				return ""
			}),
			WithUserConfigDir(func() (string, error) {
				return "/fake/config/dir", nil
			}),
		)
		expected := filepath.Join("/fake/config/dir", "goa4web", "secret1")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}
