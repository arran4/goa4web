package emailutil

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

func NotifyChange(ctx context.Context, provider email.Provider, emailAddr string, page string) error {
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

	// Define email content
	content := EmailContent{
		To:      emailAddr,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}

	// Create a new buffer to store the rendered email content
	var notification bytes.Buffer

	// Parse and execute the email template
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}

	// Execute the template and store the result in the notification buffer
	err = tmpl.Execute(&notification, content)
	if err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}

	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok {
		if err := q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToEmail: emailAddr, Subject: content.Subject, Body: notification.String()}); err != nil {
			return err
		}
	} else if provider != nil {
		if err := provider.Send(ctx, emailAddr, content.Subject, notification.String()); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	return nil
}

// getEmailProvider returns the mail provider configured by environment variables.
// Production code uses this, while tests can call email.ProviderFromConfig directly.
func getEmailProvider() email.Provider {
	return email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig)
}

// loadEmailConfigFile reads EMAIL_* style configuration values from a simple
// key=value file. Missing files return an empty configuration.

// getAdminEmails returns a slice of administrator email addresses. If the
// ADMIN_EMAILS environment variable is set, it takes precedence and is
// interpreted as a comma-separated list. If not set and a Queries value is
// provided, the database is queried for administrator accounts.
func getAdminEmails(ctx context.Context, q *db.Queries) []string {
	env := os.Getenv(config.EnvAdminEmails)
	var emails []string
	if env != "" {
		for _, e := range strings.Split(env, ",") {
			if addr := strings.TrimSpace(e); addr != "" {
				emails = append(emails, addr)
			}
		}
		return emails
	}
	if q != nil {
		rows, err := q.ListAdministratorEmails(ctx)
		if err != nil {
			log.Printf("list admin emails: %v", err)
			return emails
		}
		for _, email := range rows {
			if email.Valid {
				emails = append(emails, email.String)
			}
		}
	}
	return emails
}

// notifyAdmins sends a change notification email to all administrator
// addresses returned by getAdminEmails.
func adminNotificationsEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvAdminNotify))
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

func notifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	if provider == nil || !adminNotificationsEnabled() {
		return
	}
	for _, email := range getAdminEmails(ctx, q) {
		if err := NotifyChange(ctx, provider, email, page); err != nil {
			log.Printf("Error: NotifyChange: %s", err)
		}
	}
}

// notifyThreadSubscribers emails users subscribed to the forum thread.
func NotifyThreadSubscribers(ctx context.Context, provider email.Provider, q *db.Queries, threadID, excludeUser int32, page string) {
	if provider == nil {
		return
	}
	rows, err := q.ListUsersSubscribedToThread(ctx, db.ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: threadID,
		Idusers:                  excludeUser,
	})
	if err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
		return
	}
	for _, row := range rows {
		if err := NotifyChange(ctx, provider, row.Username.String, page); err != nil {
			log.Printf("Error: NotifyChange: %s", err)
		}
	}
}
