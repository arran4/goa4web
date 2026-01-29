package app

import (
	"context"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestNewServer_ImageSignSecretValidation(t *testing.T) {
	// Case 1: ImageSignSecret configured but not provided via options
	cfg := &config.RuntimeConfig{
		ImageSignSecret: "some_secret",
	}
	// passing a dummy session secret to avoid failing on session secret check
	_, err := NewServer(context.Background(), cfg, nil, WithSessionSecret("dummy_session"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "image signing is required but no image signing secret was configured") {
		t.Fatalf("expected error about image signing, got: %v", err)
	}
}

func TestNewServer_ImageSignSecretFileValidation(t *testing.T) {
	// Case 2: ImageSignSecretFile configured but not provided via options
	cfg := &config.RuntimeConfig{
		ImageSignSecretFile: "some_file_path",
	}
	// passing a dummy session secret to avoid failing on session secret check
	_, err := NewServer(context.Background(), cfg, nil, WithSessionSecret("dummy_session"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "image signing is required but no image signing secret was configured") {
		t.Fatalf("expected error about image signing, got: %v", err)
	}
}

func TestNewServer_ImageSignSecretProvided(t *testing.T) {
	// Case 3: ImageSignSecret configured AND provided via options
	cfg := &config.RuntimeConfig{
		ImageSignSecret: "some_secret",
	}
	_, err := NewServer(context.Background(), cfg, nil,
		WithSessionSecret("dummy_session"),
		WithImageSignSecret("some_secret"),
	)

	// We expect it to PASS the image sign check.
	// It will likely fail at DB check (PerformChecks) because we didn't mock DB fully.
	// But it must NOT fail with "image signing is required..."

	if err == nil {
		// It might fail later, but if it's nil, that's also fine for this check (though unlikely without DB)
		return
	}

	if strings.Contains(err.Error(), "image signing is required but no image signing secret was configured") {
		t.Fatalf("unexpected error about image signing when secret IS provided: %v", err)
	}
}
