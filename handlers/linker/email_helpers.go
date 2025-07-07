package linker

import (
	"bytes"
	"context"
	"fmt"
	db "github.com/arran4/goa4web/internal/db"
	"os"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr string, page string) error {
	if emailAddr == "" {
		return fmt.Errorf("no email specified")
	}
	if !emailSendingEnabled() {
		return nil
	}
	from := email.SourceEmail
	type EmailContent struct {
		To      string
		From    string
		Subject string
		URL     string
	}
	content := EmailContent{
		To:      emailAddr,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}
	var buf bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}
	if err = tmpl.Execute(&buf, content); err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok {
		if err := q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToEmail: emailAddr, Subject: content.Subject, Body: buf.String()}); err != nil {
			return err
		}
	} else if provider != nil {
		if err := provider.Send(ctx, emailAddr, content.Subject, buf.String(), ""); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	return nil
}

func emailSendingEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvEmailEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

func getUpdateEmailText(ctx context.Context) string {
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, "updateEmail"); err == nil && body != "" {
			return body
		}
	}
	return defaultUpdateEmailText
}

// defaultUpdateEmailText is the compiled-in notification template.
var defaultUpdateEmailText = templates.UpdateEmailText
