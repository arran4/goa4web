package config

import (
	"github.com/arran4/goa4web"
	"path/filepath"
	"testing"

	"github.com/arran4/goa4web/core"
)

func TestLoadOrCreateAdminAPISecretCLI(t *testing.T) {
	secret, err := LoadOrCreateAdminAPISecret(core.UseMemFS(t), "cli", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "cli" {
		t.Fatalf("want cli got %q", secret)
	}
}

func TestLoadOrCreateAdminAPISecretEnv(t *testing.T) {
	t.Setenv(EnvAdminAPISecret, "env")
	secret, err := LoadOrCreateAdminAPISecret(core.UseMemFS(t), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "env" {
		t.Fatalf("want env got %q", secret)
	}
}

func TestLoadOrCreateAdminAPISecretFile(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "sec"
	if err := fs.WriteFile(file, []byte("fromfile"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	secret, err := LoadOrCreateAdminAPISecret(fs, "", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "fromfile" {
		t.Fatalf("want fromfile got %q", secret)
	}
}

func TestLoadOrCreateAdminAPISecretGenerate(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "new"
	secret, err := LoadOrCreateAdminAPISecret(fs, "", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret == "" {
		t.Fatal("secret should not be empty")
	}
	b, err := fs.ReadFile(file)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(b) != secret {
		t.Fatalf("file secret mismatch: %s vs %s", b, secret)
	}
}

func TestDefaultAdminAPISecretPathDev(t *testing.T) {
	goa4web.Version = "dev"
	t.Setenv(EnvDocker, "")
	got := DefaultAdminAPISecretPath()
	if got != ".admin_api_secret" {
		t.Fatalf("want .admin_api_secret got %s", got)
	}
}

func TestDefaultAdminAPISecretPathDocker(t *testing.T) {
	goa4web.Version = "1"
	t.Setenv(EnvDocker, "1")
	t.Setenv("HOME", "/home/test")
	got := DefaultAdminAPISecretPath()
	if got != "/var/lib/goa4web/admin_api_secret" {
		t.Fatalf("want /var/lib/goa4web/admin_api_secret got %s", got)
	}
}

func TestDefaultAdminAPISecretPathUser(t *testing.T) {
	goa4web.Version = "1"
	t.Setenv(EnvDocker, "")
	t.Setenv("XDG_CONFIG_HOME", "/cfg")
	got := DefaultAdminAPISecretPath()
	want := filepath.Join("/cfg", "goa4web", "admin_api_secret")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}
