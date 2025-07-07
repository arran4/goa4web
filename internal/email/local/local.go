package local

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider relies on the local sendmail binary.
type Provider struct{}

func (Provider) Send(ctx context.Context, to, subject string, rawEmailMessage []byte) error {
	cmd := exec.CommandContext(ctx, "sendmail", to)
	cmd.Stdin = bytes.NewReader(rawEmailMessage)
	return cmd.Run()
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return Provider{} }

// Register registers the local provider factory.
func Register() { email.RegisterProvider("local", providerFromConfig) }
