//go:build ses
// +build ses

package emaildefaults_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	sesProv "github.com/arran4/goa4web/internal/email/ses"
)

func init() {
	email.DefaultRegistry = email.NewRegistry()
	sesProv.Register(email.DefaultRegistry)
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	if p := email.ProviderFromConfig(config.RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}
