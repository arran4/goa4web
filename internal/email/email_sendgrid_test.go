//go:build sendgrid
// +build sendgrid

package email_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	sendgridProv "github.com/arran4/goa4web/internal/email/sendgrid"
)

func init() {
	email.DefaultRegistry = email.NewRegistry()
	sendgridProv.Register(email.DefaultRegistry)
}

func TestSendGridProviderFromConfig(t *testing.T) {
	p := ProviderFromConfig(config.RuntimeConfig{EmailProvider: "sendgrid", EmailSendGridKey: "k", EmailFrom: "from@example.com"})
	if _, ok := p.(sendgridProv.Provider); !ok {
		t.Fatalf("expected SendGridProvider, got %#v", p)
	}
}
