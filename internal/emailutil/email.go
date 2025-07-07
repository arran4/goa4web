package emailutil

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/goa4web/config"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

type emailTemplate struct {
	text string
	html string
}

var defaultEmailTemplates = map[string]emailTemplate{
	"update":                                   {text: defaultUpdateEmailText, html: defaultUpdateEmailHTML},
	strings.ToLower(hcommon.TaskReply):         {text: defaultReplyEmailText, html: defaultReplyEmailHTML},
	strings.ToLower(hcommon.TaskCreateThread):  {text: defaultThreadEmailText, html: defaultThreadEmailHTML},
	strings.ToLower(hcommon.TaskNewPost):       {text: defaultBlogEmailText, html: defaultBlogEmailHTML},
	strings.ToLower(hcommon.TaskSubmitWriting): {text: defaultWritingEmailText, html: defaultWritingEmailHTML},
	strings.ToLower(hcommon.TaskRegister):      {text: defaultSignupEmailText, html: defaultSignupEmailHTML},
}

func getEmailTemplates(ctx context.Context, action string) (string, string) {
	name := "email_" + strings.ToLower(action)
	nameHTML := name + "_html"
	var text, html string
	if q, ok := ctx.Value(hcommon.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, name); err == nil && body != "" {
			text = body
		}
		if body, err := q.GetTemplateOverride(ctx, nameHTML); err == nil && body != "" {
			html = body
		}
	}
	if t, ok := defaultEmailTemplates[strings.ToLower(action)]; ok {
		if text == "" {
			text = t.text
		}
		if html == "" {
			html = t.html
		}
	}
	return text, html
}

func NotifyChange(ctx context.Context, provider email.Provider, emailAddr, page, action string, item interface{}) error {
	if emailAddr == "" {
		return fmt.Errorf("no email specified")
	}
	if !EmailSendingEnabled() {
		return nil
	}
	from := runtimeconfig.AppRuntimeConfig.EmailFrom

	type EmailContent struct {
		To      string
		From    string
		Subject string
		URL     string
		Action  string
		Path    string
		Time    string
		Item    interface{}
	}

	// Define email content
	content := EmailContent{
		To:      emailAddr,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
		Action:  action,
		Path:    page,
		Time:    time.Now().Format(time.RFC822),
		Item:    item,
	}

	// Create a new buffer to store the rendered email content
	var textBody, htmlBody string
	tmplText, tmplHTML := getEmailTemplates(ctx, action)
	if tmplText == "" && tmplHTML == "" {
		return nil
	}
	if tmplText != "" {
		var buf bytes.Buffer
		t, err := template.New("text").Parse(tmplText)
		if err != nil {
			return fmt.Errorf("parse email template: %w", err)
		}
		if err := t.Execute(&buf, content); err != nil {
			return fmt.Errorf("execute email template: %w", err)
		}
		textBody = buf.String()
	}
	if tmplHTML != "" {
		var buf bytes.Buffer
		t, err := template.New("html").Parse(tmplHTML)
		if err != nil {
			return fmt.Errorf("parse email html template: %w", err)
		}
		if err := t.Execute(&buf, content); err != nil {
			return fmt.Errorf("execute email html template: %w", err)
		}
		htmlBody = buf.String()
	}

	if q, ok := ctx.Value(hcommon.KeyQueries).(*db.Queries); ok {
		if err := q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToEmail: emailAddr, Subject: content.Subject, Body: textBody, HtmlBody: htmlBody}); err != nil {
			return err
		}
	} else if provider != nil {
		if err := provider.Send(ctx, emailAddr, content.Subject, textBody, htmlBody); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	return nil
}

// getEmailProvider returns the mail provider configured by environment variables.
// Production code uses this, while tests can call email.ProviderFromConfig directly.

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

// AdminNotificationsEnabled reports whether administrator notification emails
// should be sent. The ADMIN_NOTIFY environment variable can be set to any of
// "0", "false", "off" or "no" to disable notifications.
func AdminNotificationsEnabled() bool {
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

func EmailSendingEnabled() bool {
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

// NotifyAdmins sends a change notification email to all administrator addresses
// returned by GetAdminEmails.
func NotifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	if provider == nil || !AdminNotificationsEnabled() {
		return
	}
	for _, email := range GetAdminEmails(ctx, q) {
		if err := NotifyChange(ctx, provider, email, page, "update", nil); err != nil {
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
		if err := NotifyChange(ctx, provider, row.Username.String, page, "update", nil); err != nil {
			log.Printf("Error: NotifyChange: %s", err)
		}
	}
}
