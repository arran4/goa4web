package local

import (
	"bytes"
	"context"
	"net/mail"
	"os/exec"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider relies on the local sendmail binary.
type Provider struct{}

func (Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	cmd := exec.CommandContext(ctx, "sendmail", to.Address)
	cmd.Stdin = bytes.NewReader(rawEmailMessage)
	return cmd.Run()
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return Provider{} }

// Register registers the local provider factory.
func Register() { email.RegisterProvider("local", providerFromConfig) }
