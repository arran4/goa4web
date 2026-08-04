package secrets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arran4/goa4web"
)

func TestDefaultPath(t *testing.T) {
	origVersion := goa4web.Version
	t.Cleanup(func() {
		goa4web.Version = origVersion
	})

	t.Run("dev version", func(t *testing.T) {
		goa4web.Version = "dev"
		t.Setenv("MY_DOCKER_ENV", "1")
		got := DefaultPath("secret1", "MY_DOCKER_ENV")
		if got != ".secret1" {
			t.Errorf("expected %q, got %q", ".secret1", got)
		}
	})

	t.Run("docker env set", func(t *testing.T) {
		goa4web.Version = "v1.0.0"
		t.Setenv("MY_DOCKER_ENV", "1")
		got := DefaultPath("secret1", "MY_DOCKER_ENV")
		expected := filepath.Join("/var/lib/goa4web", "secret1")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("no home or xdg", func(t *testing.T) {
		goa4web.Version = "v1.0.0"
		t.Setenv("HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "")
		got := DefaultPath("secret1", "MISSING_DOCKER_ENV")
		expected := filepath.Join("/var/lib/goa4web", "secret1")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("user config dir", func(t *testing.T) {
		goa4web.Version = "v1.0.0"
		t.Setenv("HOME", "/fake/home")

		dir, err := os.UserConfigDir()
		if err != nil {
			t.Skip("os.UserConfigDir() returned an error, skipping test")
		}

		got := DefaultPath("secret1", "MISSING_DOCKER_ENV")
		expected := filepath.Join(dir, "goa4web", "secret1")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}
