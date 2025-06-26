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

// NotifyChange sends a simple page update email to emailAddr. If
// EMAIL_ENABLED is set to a false value the notification is skipped. When the
// context contains a *db.Queries value the message is queued instead of sent
// immediately. If neither a queue nor provider is available the call is a no-op.
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
// GetAdminEmails returns a slice of administrator email addresses. If the
// ADMIN_EMAILS environment variable is set, it takes precedence and is
// interpreted as a comma-separated list. If not set and a Queries value is
// provided, the database is queried for administrator accounts.
func GetAdminEmails(ctx context.Context, q *db.Queries) []string {
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

<<<<<<< i3595l-codex/refactor-duplicated-functions-in-email_helpers
// AdminNotificationsEnabled reports whether administrator notification emails
// should be sent. The ADMIN_NOTIFY environment variable can be set to any of
// "0", "false", "off" or "no" to disable notifications.
func AdminNotificationsEnabled() bool {
=======
// adminNotificationsEnabled reports whether administrator notifications should
// be delivered. Setting ADMIN_NOTIFY to "0", "false", "off" or "no" disables
// these messages.
func adminNotificationsEnabled() bool {
>>>>>>> main
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

// emailSendingEnabled reports whether the application should attempt to send
// queued emails. A false value for EMAIL_ENABLED ("0", "false", "off" or "no")
// disables sending entirely.
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

<<<<<<< i3595l-codex/refactor-duplicated-functions-in-email_helpers
// NotifyAdmins sends a change notification email to all administrator addresses
// returned by GetAdminEmails.
func NotifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	if provider == nil || !AdminNotificationsEnabled() {
=======
// notifyAdmins emails a page update to all administrator addresses returned by
// getAdminEmails. The notification is skipped when no provider is configured or
// ADMIN_NOTIFY disables administrator messages.
func notifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	if provider == nil || !adminNotificationsEnabled() {
>>>>>>> main
		return
	}
	for _, email := range GetAdminEmails(ctx, q) {
		if err := NotifyChange(ctx, provider, email, page); err != nil {
			log.Printf("Error: NotifyChange: %s", err)
		}
	}
}

// NotifyThreadSubscribers emails users subscribed to the forum thread. If the
// provider is nil or the query fails, no notifications are sent. Each call to
// NotifyChange respects EMAIL_ENABLED when delivering the message.
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
