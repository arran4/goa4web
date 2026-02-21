package local

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"os/exec"
	"unicode"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Provider relies on the local sendmail binary.
type Provider struct{}

func (Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	addr := to.Address
	if addr == "" {
		return fmt.Errorf("recipient address empty")
	}
	for _, r := range addr {
		if r == '\n' || r == '\r' || unicode.IsControl(r) {
			return fmt.Errorf("invalid recipient address")
		}
	}
	parsed, err := mail.ParseAddress(addr)
	if err != nil || parsed.Address != addr {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	cmd := exec.CommandContext(ctx, "sendmail", addr)
	cmd.Stdin = bytes.NewReader(rawEmailMessage)
	return cmd.Run()
}

// TestConfig checks if the sendmail binary is available.
func (Provider) TestConfig(ctx context.Context) (string, error) {
	_, err := exec.LookPath("sendmail")
	if err != nil {
		return "", fmt.Errorf("sendmail binary not found: %w", err)
	}
	return "sendmail binary found", nil
}

func providerFromConfig(*config.RuntimeConfig) (email.Provider, error) { return Provider{}, nil }

// Register registers the local provider factory.
func Register(r *email.Registry) { r.RegisterProvider("local", providerFromConfig) }
