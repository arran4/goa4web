//go:build ses
// +build ses

package emaildefaults_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	sesProv "github.com/arran4/goa4web/internal/email/ses"
)

func newRegistry() *email.Registry {
	r := email.NewRegistry()
	sesProv.Register(r)
	return r
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	reg := newRegistry()
	if p := reg.ProviderFromConfig(&config.RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}
