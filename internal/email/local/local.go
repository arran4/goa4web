package local

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider relies on the local sendmail binary.
type Provider struct{}

func (Provider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	cmd := exec.CommandContext(ctx, "sendmail", to)
	body := textBody
	if htmlBody != "" {
		boundary := "a4web" + strings.ReplaceAll(fmt.Sprint(time.Now().UnixNano()), "-", "")
		buf := strings.Builder{}
		fmt.Fprintf(&buf, "Subject: %s\nMIME-Version: 1.0\nContent-Type: multipart/alternative; boundary=%s\n\n", subject, boundary)
		fmt.Fprintf(&buf, "--%s\nContent-Type: text/plain; charset=utf-8\n\n%s\n", boundary, textBody)
		fmt.Fprintf(&buf, "--%s\nContent-Type: text/html; charset=utf-8\n\n%s\n--%s--", boundary, htmlBody, boundary)
		body = buf.String()
		cmd.Stdin = strings.NewReader(body)
	} else {
		cmd.Stdin = strings.NewReader("Subject: " + subject + "\n\n" + body)
	}
	return cmd.Run()
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return Provider{} }

// Register registers the local provider factory.
func Register() { email.RegisterProvider("local", providerFromConfig) }
