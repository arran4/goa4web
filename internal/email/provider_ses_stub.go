//go:build !ses
// +build !ses

package email

import (
	"context"

	"github.com/arran4/goa4web/runtimeconfig"
)

// SESBuilt indicates whether the SES provider is compiled in.
const SESBuilt = false

// SESProvider is a stub implementation used when SES support is disabled.
type SESProvider struct{}

// Send implements the Provider interface but performs no action.
func (SESProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	return nil
}

// SESProviderFromConfig returns nil when SES support is disabled.
func SESProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider { return nil }
