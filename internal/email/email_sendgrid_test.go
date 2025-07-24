//go:build sendgrid
// +build sendgrid

package email_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	sendgridProv "github.com/arran4/goa4web/internal/email/sendgrid"
)

func newRegistry() *email.Registry {
	r := email.NewRegistry()
	sendgridProv.Register(r)
	return r
}

func TestSendGridProviderFromConfig(t *testing.T) {
	reg := newRegistry()
	p := reg.ProviderFromConfig(config.RuntimeConfig{EmailProvider: "sendgrid", EmailSendGridKey: "k", EmailFrom: "from@example.com"})
	if _, ok := p.(sendgridProv.Provider); !ok {
		t.Fatalf("expected SendGridProvider, got %#v", p)
	}
}
