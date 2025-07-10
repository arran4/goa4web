//go:build sendgrid
// +build sendgrid

package emailutil_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	sendgridProv "github.com/arran4/goa4web/internal/email/sendgrid"
)

func init() {
	sendgridProv.Register()
}

func TestSendGridProviderFromConfig(t *testing.T) {
	p := email.ProviderFromConfig(config.RuntimeConfig{EmailProvider: "sendgrid", EmailSendGridKey: "k", EmailFrom: "from@example.com"})
	if _, ok := p.(sendgridProv.Provider); !ok {
		t.Fatalf("expected SendGridProvider, got %#v", p)
	}
}
