package app

import (
	"testing"
)

func TestWithImageSignSecret(t *testing.T) {
	secret := "test-image-secret"
	opt := WithImageSignSecret(secret)

	opts := &serverOptions{}
	opt(opts)

	if opts.ImageSignSecret != secret {
		t.Errorf("expected ImageSignSecret to be %q, got %q", secret, opts.ImageSignSecret)
	}
}
