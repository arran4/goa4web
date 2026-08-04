package app

import (
	"testing"
)

func TestWithSessionSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedSecret := "test_secret_123"

	option := WithSessionSecret(expectedSecret)
	option(opts)

	if opts.SessionSecret != expectedSecret {
		t.Errorf("WithSessionSecret() = %v, want %v", opts.SessionSecret, expectedSecret)
	}
}

func TestWithImageSignSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedSecret := "img_secret_456"

	option := WithImageSignSecret(expectedSecret)
	option(opts)

	if opts.ImageSignSecret != expectedSecret {
		t.Errorf("WithImageSignSecret() = %v, want %v", opts.ImageSignSecret, expectedSecret)
	}
}

func TestWithLinkSignSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedSecret := "link_secret_789"

	option := WithLinkSignSecret(expectedSecret)
	option(opts)

	if opts.LinkSignSecret != expectedSecret {
		t.Errorf("WithLinkSignSecret() = %v, want %v", opts.LinkSignSecret, expectedSecret)
	}
}

func TestWithShareSignSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedSecret := "share_secret_101"

	option := WithShareSignSecret(expectedSecret)
	option(opts)

	if opts.ShareSignSecret != expectedSecret {
		t.Errorf("WithShareSignSecret() = %v, want %v", opts.ShareSignSecret, expectedSecret)
	}
}

func TestWithAPISecret(t *testing.T) {
	opts := &serverOptions{}
	expectedSecret := "api_secret_202"

	option := WithAPISecret(expectedSecret)
	option(opts)

	if opts.APISecret != expectedSecret {
		t.Errorf("WithAPISecret() = %v, want %v", opts.APISecret, expectedSecret)
	}
}
