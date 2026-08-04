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

func TestWithLinkSignSecret(t *testing.T) {
	tests := []struct {
		name          string
		initialSecret string
		newSecret     string
	}{
		{
			name:          "empty initial secret",
			initialSecret: "",
			newSecret:     "test-link-sign-secret",
		},
		{
			name:          "overwrite existing secret",
			initialSecret: "old-secret",
			newSecret:     "new-secret",
		},
		{
			name:          "set to empty secret",
			initialSecret: "existing-secret",
			newSecret:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &serverOptions{
				LinkSignSecret: tt.initialSecret,
			}

			opt := WithLinkSignSecret(tt.newSecret)
			opt(o)

			if o.LinkSignSecret != tt.newSecret {
				t.Errorf("expected LinkSignSecret to be %q, got %q", tt.newSecret, o.LinkSignSecret)
			}
		})
	}
}
