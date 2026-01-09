package config

import (
	"path/filepath"
	"testing"

	"github.com/arran4/goa4web"

	"github.com/arran4/goa4web/core"
)

func TestLoadOrCreateSecretCLI(t *testing.T) {
	secret, err := LoadOrCreateSessionSecret(core.UseMemFS(t), "cli", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "cli" {
		t.Fatalf("want cli got %q", secret)
	}
}

func TestLoadOrCreateSecretEnv(t *testing.T) {
	t.Setenv(EnvSessionSecret, "env")
	secret, err := LoadOrCreateSessionSecret(core.UseMemFS(t), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "env" {
		t.Fatalf("want env got %q", secret)
	}
}

func TestLoadOrCreateSecretFile(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "sec"
	if err := fs.WriteFile(file, []byte("fromfile"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	secret, err := LoadOrCreateSessionSecret(fs, "", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "fromfile" {
		t.Fatalf("want fromfile got %q", secret)
	}
}

func TestLoadOrCreateSecretGenerate(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "new"
	secret, err := LoadOrCreateSessionSecret(fs, "", file)
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

func TestDefaultSessionSecretPathDev(t *testing.T) {
	goa4web.Version = "dev"
	t.Setenv(EnvDocker, "")
	got := DefaultSessionSecretPath()
	if got != ".session_secret" {
		t.Fatalf("want .session_secret got %s", got)
	}
}

func TestDefaultSessionSecretPathDocker(t *testing.T) {
	goa4web.Version = "1"
	t.Setenv(EnvDocker, "1")
	t.Setenv("HOME", "/home/test")
	got := DefaultSessionSecretPath()
	if got != "/var/lib/goa4web/session_secret" {
		t.Fatalf("want /var/lib/goa4web/session_secret got %s", got)
	}
}

func TestDefaultSessionSecretPathUser(t *testing.T) {
	goa4web.Version = "1"
	t.Setenv(EnvDocker, "")
	t.Setenv("XDG_CONFIG_HOME", "/cfg")
	got := DefaultSessionSecretPath()
	want := filepath.Join("/cfg", "goa4web", "session_secret")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}
