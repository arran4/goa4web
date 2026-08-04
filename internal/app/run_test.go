package app

import (
	"testing"
)

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
