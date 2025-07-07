package core_test

import (
	"path/filepath"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestLoadSessionSecretCLI(t *testing.T) {
	secret, err := core.LoadSessionSecret(core.OSFS{}, "cli", "", config.EnvSessionSecret, config.EnvSessionSecretFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "cli" {
		t.Fatalf("want cli got %q", secret)
	}
}

func TestLoadSessionSecretEnv(t *testing.T) {
	t.Setenv(config.EnvSessionSecret, "env")
	secret, err := core.LoadSessionSecret(core.OSFS{}, "", "", config.EnvSessionSecret, config.EnvSessionSecretFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "env" {
		t.Fatalf("want env got %q", secret)
	}
}

func TestLoadSessionSecretFile(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "sec"
	if err := fs.WriteFile(file, []byte("fromfile"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	secret, err := core.LoadSessionSecret(fs, "", file, config.EnvSessionSecret, config.EnvSessionSecretFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "fromfile" {
		t.Fatalf("want fromfile got %q", secret)
	}
}

func TestLoadSessionSecretGenerate(t *testing.T) {
	fs := core.UseMemFS(t)
	file := "new"
	secret, err := core.LoadSessionSecret(fs, "", file, config.EnvSessionSecret, config.EnvSessionSecretFile)
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
	common.Version = "dev"
	t.Setenv(config.EnvDocker, "")
	got := runtimeconfig.DefaultSessionSecretPath()
	if got != ".session_secret" {
		t.Fatalf("want .session_secret got %s", got)
	}
}

func TestDefaultSessionSecretPathDocker(t *testing.T) {
	common.Version = "1"
	t.Setenv(config.EnvDocker, "1")
	t.Setenv("HOME", "/home/test")
	got := runtimeconfig.DefaultSessionSecretPath()
	if got != "/var/lib/goa4web/session_secret" {
		t.Fatalf("want /var/lib/goa4web/session_secret got %s", got)
	}
}

func TestDefaultSessionSecretPathUser(t *testing.T) {
	common.Version = "1"
	t.Setenv(config.EnvDocker, "")
	t.Setenv("XDG_CONFIG_HOME", "/cfg")
	got := runtimeconfig.DefaultSessionSecretPath()
	want := filepath.Join("/cfg", "goa4web", "session_secret")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}
