package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestApplySMTPFallbacksUseUser(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "smtp", EmailSMTPUser: "user@example.com"}
	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if cfg.EmailFrom != "user@example.com" {
		t.Fatalf("from=%q", cfg.EmailFrom)
	}
}

func TestApplySMTPFallbacksUseFrom(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "smtp", EmailFrom: "from@example.com"}
	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if cfg.EmailSMTPUser != "from@example.com" {
		t.Fatalf("user=%q", cfg.EmailSMTPUser)
	}
}

func TestApplySMTPFallbacksBothBlank(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "smtp"}
	if err := config.ApplySMTPFallbacks(&cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestApplySMTPFallbacksNoSMTP(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "log"}
	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		t.Fatalf("apply: %v", err)
	}
}

func TestApplySMTPFallbacksBothSet(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "smtp", EmailFrom: "a@example.com", EmailSMTPUser: "user"}
	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if cfg.EmailFrom != "a@example.com" || cfg.EmailSMTPUser != "user" {
		t.Fatalf("changed: %#v", cfg)
	}
}

func TestApplySMTPFallbacksMismatch(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "smtp", EmailFrom: "a@example.com", EmailSMTPUser: "b@example.com"}
	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if cfg.EmailFrom != "a@example.com" || cfg.EmailSMTPUser != "b@example.com" {
		t.Fatalf("changed: %#v", cfg)
	}
}
