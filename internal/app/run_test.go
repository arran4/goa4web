package app

import (
	"testing"
)

func TestWithSessionSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedValue := "dummy-session-value-1"

	option := WithSessionSecret(expectedValue)
	option(opts)

	if opts.SessionSecret != expectedValue {
		t.Errorf("WithSessionSecret() = %v, want %v", opts.SessionSecret, expectedValue)
	}
}

func TestWithImageSignSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedValue := "dummy-image-value-2"

	option := WithImageSignSecret(expectedValue)
	option(opts)

	if opts.ImageSignSecret != expectedValue {
		t.Errorf("WithImageSignSecret() = %v, want %v", opts.ImageSignSecret, expectedValue)
	}
}

func TestWithLinkSignSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedValue := "dummy-link-value-3"

	option := WithLinkSignSecret(expectedValue)
	option(opts)

	if opts.LinkSignSecret != expectedValue {
		t.Errorf("WithLinkSignSecret() = %v, want %v", opts.LinkSignSecret, expectedValue)
	}
}

func TestWithShareSignSecret(t *testing.T) {
	opts := &serverOptions{}
	expectedValue := "dummy-share-value-4"

	option := WithShareSignSecret(expectedValue)
	option(opts)

	if opts.ShareSignSecret != expectedValue {
		t.Errorf("WithShareSignSecret() = %v, want %v", opts.ShareSignSecret, expectedValue)
	}
}

func TestWithAPISecret(t *testing.T) {
	opts := &serverOptions{}
	expectedValue := "dummy-api-value-5"

	option := WithAPISecret(expectedValue)
	option(opts)

	if opts.APISecret != expectedValue {
		t.Errorf("WithAPISecret() = %v, want %v", opts.APISecret, expectedValue)
	}
}
