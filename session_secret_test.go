package goa4web

import "testing"

func TestLoadSessionSecretCLI(t *testing.T) {
	secret, err := LoadSessionSecret("cli", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "cli" {
		t.Fatalf("want cli got %q", secret)
	}
}

func TestLoadSessionSecretEnv(t *testing.T) {
	t.Setenv("SESSION_SECRET", "env")
	secret, err := LoadSessionSecret("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "env" {
		t.Fatalf("want env got %q", secret)
	}
}

func TestLoadSessionSecretFile(t *testing.T) {
	useMemFS(t)
	file := "sec"
	if err := writeFile(file, []byte("fromfile"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	secret, err := LoadSessionSecret("", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "fromfile" {
		t.Fatalf("want fromfile got %q", secret)
	}
}

func TestLoadSessionSecretGenerate(t *testing.T) {
	fs := useMemFS(t)
	file := "new"
	secret, err := LoadSessionSecret("", file)
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
