package app

import (
	"testing"
)

func TestWithShareSignSecret(t *testing.T) {
	opts := &serverOptions{}
	opt := WithShareSignSecret("my-secret")
	opt(opts)

	if opts.ShareSignSecret != "my-secret" {
		t.Errorf("expected ShareSignSecret to be 'my-secret', got '%s'", opts.ShareSignSecret)
	}
}
