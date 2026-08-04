package app

import (
	"testing"
)

func TestWithLinkSignSecret(t *testing.T) {
	secret := "test-link-sign-secret"
	opt := WithLinkSignSecret(secret)

	o := &serverOptions{}
	opt(o)

	if o.LinkSignSecret != secret {
		t.Errorf("expected LinkSignSecret to be %q, got %q", secret, o.LinkSignSecret)
	}
}
