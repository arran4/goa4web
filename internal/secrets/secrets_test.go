package secrets

import (
	"errors"
	"io/fs"
	"testing"
	"encoding/hex"

	"github.com/arran4/goa4web/core"
)

type errorFS struct {
	readErr  error
	writeErr error
	memFS    core.FileSystem
}

func (e *errorFS) ReadFile(name string) ([]byte, error) {
	if e.readErr != nil {
		return nil, e.readErr
	}
	return e.memFS.ReadFile(name)
}

func (e *errorFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if e.writeErr != nil {
		return e.writeErr
	}
	return e.memFS.WriteFile(name, data, perm)
}

func TestLoadOrCreate(t *testing.T) {
	defaultPathFunc := func() string { return "default_path_secret.txt" }

	t.Run("cliSecret takes precedence", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		secret, err := LoadOrCreate(memFs, "my_cli_secret", "", "", "", defaultPathFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secret != "my_cli_secret" {
			t.Errorf("expected secret %q, got %q", "my_cli_secret", secret)
		}
	})

	t.Run("envSecret takes precedence over file", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		t.Setenv("TEST_ENV_SECRET", "my_env_secret")
		secret, err := LoadOrCreate(memFs, "", "", "TEST_ENV_SECRET", "", defaultPathFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secret != "my_env_secret" {
			t.Errorf("expected secret %q, got %q", "my_env_secret", secret)
		}
	})

	t.Run("loads from path", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		if err := memFs.WriteFile("my_path_secret.txt", []byte("my_path_secret"), 0600); err != nil {
			t.Fatalf("failed to write mock file: %v", err)
		}
		secret, err := LoadOrCreate(memFs, "", "my_path_secret.txt", "", "", defaultPathFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secret != "my_path_secret" {
			t.Errorf("expected secret %q, got %q", "my_path_secret", secret)
		}
	})

	t.Run("falls back to envSecretFile", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		t.Setenv("TEST_ENV_SECRET_FILE", "my_env_path_secret.txt")
		if err := memFs.WriteFile("my_env_path_secret.txt", []byte("my_env_path_secret"), 0600); err != nil {
			t.Fatalf("failed to write mock file: %v", err)
		}
		secret, err := LoadOrCreate(memFs, "", "", "", "TEST_ENV_SECRET_FILE", defaultPathFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secret != "my_env_path_secret" {
			t.Errorf("expected secret %q, got %q", "my_env_path_secret", secret)
		}
	})

	t.Run("falls back to defaultPath", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		if err := memFs.WriteFile("default_path_secret.txt", []byte("my_default_path_secret"), 0600); err != nil {
			t.Fatalf("failed to write mock file: %v", err)
		}
		secret, err := LoadOrCreate(memFs, "", "", "", "", defaultPathFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secret != "my_default_path_secret" {
			t.Errorf("expected secret %q, got %q", "my_default_path_secret", secret)
		}
	})

	t.Run("generates new secret if file does not exist", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		path := "new_secret.txt"
		secret, err := LoadOrCreate(memFs, "", path, "", "", defaultPathFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(secret) != 64 {
			t.Errorf("expected 64 character hex string, got %d characters: %q", len(secret), secret)
		}

		if _, err := hex.DecodeString(secret); err != nil {
			t.Errorf("expected valid hex string, got error: %v", err)
		}

		b, err := memFs.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file to be created, got error: %v", err)
		}
		if string(b) != secret {
			t.Errorf("expected file contents %q, got %q", secret, string(b))
		}
	})

	t.Run("returns error on read failure (non-NotExist)", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		expectedErr := errors.New("read error")
		errFs := &errorFS{readErr: expectedErr, memFS: memFs}

		_, err := LoadOrCreate(errFs, "", "some_path.txt", "", "", defaultPathFunc)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("returns error on write failure", func(t *testing.T) {
		memFs := core.UseMemFS(t)
		expectedErr := errors.New("write error")
		errFs := &errorFS{writeErr: expectedErr, memFS: memFs}

		_, err := LoadOrCreate(errFs, "", "some_path.txt", "", "", defaultPathFunc)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
