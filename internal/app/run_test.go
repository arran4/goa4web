package app

import (
	"testing"
)

func TestWithShareSignSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"empty secret", ""},
		{"simple secret", "my-secret"},
		{"complex secret", "a very long and complex secret with special characters !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &serverOptions{}
			opt := WithShareSignSecret(tt.secret)
			opt(opts)

			if opts.ShareSignSecret != tt.secret {
				t.Errorf("expected ShareSignSecret to be '%s', got '%s'", tt.secret, opts.ShareSignSecret)
			}
		})
	}
}
