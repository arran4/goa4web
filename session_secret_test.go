package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSessionSecretCLI(t *testing.T) {
	secret, err := loadSessionSecret("cli", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "cli" {
		t.Fatalf("want cli got %q", secret)
	}
}

func TestLoadSessionSecretEnv(t *testing.T) {
	t.Setenv("SESSION_SECRET", "env")
	secret, err := loadSessionSecret("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "env" {
		t.Fatalf("want env got %q", secret)
	}
}

func TestLoadSessionSecretFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sec")
	if err := os.WriteFile(file, []byte("fromfile"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	secret, err := loadSessionSecret("", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret != "fromfile" {
		t.Fatalf("want fromfile got %q", secret)
	}
}

func TestLoadSessionSecretGenerate(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "new")
	secret, err := loadSessionSecret("", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secret == "" {
		t.Fatal("secret should not be empty")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(b) != secret {
		t.Fatalf("file secret mismatch: %s vs %s", b, secret)
	}
}
