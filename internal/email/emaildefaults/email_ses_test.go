//go:build ses

package emaildefaults_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	sesProv "github.com/arran4/goa4web/internal/email/ses"
)

func newSesRegistry() *email.Registry {
	r := email.NewRegistry()
	sesProv.Register(r)
	return r
}

func TestGetEmailProviderSESNoCreds_SES(t *testing.T) {
	reg := newSesRegistry()
	p, err := reg.ProviderFromConfig(&config.RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"})
	if err != nil {
		// Error is expected if credentials are missing
	}
	if p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}
