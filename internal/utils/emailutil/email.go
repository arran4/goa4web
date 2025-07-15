package emailutil

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/goa4web/config"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

type emailTemplate struct {
	text string
	html string
}

var defaultEmailTemplates = map[string]emailTemplate{
	"update":                                           {text: defaultUpdateEmailText, html: defaultUpdateEmailHTML},
	strings.ToLower(hcommon.TaskReply):                 {text: defaultReplyEmailText, html: defaultReplyEmailHTML},
	strings.ToLower(hcommon.TaskCreateThread):          {text: defaultThreadEmailText, html: defaultThreadEmailHTML},
	strings.ToLower(hcommon.TaskNewPost):               {text: defaultBlogEmailText, html: defaultBlogEmailHTML},
	strings.ToLower(hcommon.TaskSubmitWriting):         {text: defaultWritingEmailText, html: defaultWritingEmailHTML},
	strings.ToLower(hcommon.TaskRegister):              {text: defaultSignupEmailText, html: defaultSignupEmailHTML},
	strings.ToLower(hcommon.TaskUserEmailVerification): {text: defaultVerificationEmailText, html: defaultVerificationEmailHTML},
	strings.ToLower(hcommon.TaskUserResetPassword):     {text: defaultPasswordResetEmailText, html: defaultPasswordResetEmailHTML},
	"user approved":                                    {text: defaultUserApprovedEmailText, html: defaultUserApprovedEmailHTML},
	"user rejected":                                    {text: defaultUserRejectedEmailText, html: defaultUserRejectedEmailHTML},
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

func CreateEmailTemplate(ctx context.Context, emailAddr, page, action string, item interface{}) ([]byte, mail.Address, error) {
	if emailAddr == "" {
		return nil, mail.Address{}, fmt.Errorf("no email specified")
	}
	from := email.ParseAddress(config.AppRuntimeConfig.EmailFrom)

	type EmailContent struct {
		To       string
		From     string
		Subject  string
		URL      string
		Action   string
		Path     string
		Time     string
		UnsubURL string
		Item     interface{}
	}

	// Define email content
	unsub := "/usr/subscriptions"
	if config.AppRuntimeConfig.HTTPHostname != "" {
		unsub = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + unsub
	}
	toAddr := email.ParseAddress(emailAddr)
	content := EmailContent{
		To:       emailAddr,
		From:     from.Address,
		Subject:  "Website Update Notification",
		URL:      page,
		Action:   action,
		Path:     page,
		Time:     time.Now().Format(time.RFC822),
		UnsubURL: unsub,
		Item:     item,
	}

	// Create a new buffer to store the rendered email content
	var textBody, htmlBody string
	tmplText, tmplHTML := getEmailTemplates(ctx, action)
	if tmplText == "" && tmplHTML == "" {
		return nil, mail.Address{}, nil
	}
	if tmplText != "" {
		var buf bytes.Buffer
		t, err := template.New("text").Parse(tmplText)
		if err != nil {
			return nil, mail.Address{}, fmt.Errorf("parse email template: %w", err)
		}
		if err := t.Execute(&buf, content); err != nil {
			return nil, mail.Address{}, fmt.Errorf("execute email template: %w", err)
		}
		textBody = buf.String()
	}
	if tmplHTML != "" {
		var buf bytes.Buffer
		t, err := template.New("html").Parse(tmplHTML)
		if err != nil {
			return nil, mail.Address{}, fmt.Errorf("parse email html template: %w", err)
		}
		if err := t.Execute(&buf, content); err != nil {
			return nil, mail.Address{}, fmt.Errorf("execute email html template: %w", err)
		}
		htmlBody = buf.String()
	}

	msg, err := email.BuildMessage(from, toAddr, content.Subject, textBody, htmlBody)
	if err != nil {
		return nil, mail.Address{}, fmt.Errorf("build message: %w", err)
	}
	return msg, toAddr, nil
}

func CreateEmailTemplateAndQueue(ctx context.Context, q *db.Queries, userID int32, emailAddr, page, action string, item interface{}) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	msg, _, err := CreateEmailTemplate(ctx, emailAddr, page, action, item)
	if err != nil {
		return err
	}
	return q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: userID, Body: string(msg)})
}

// getEmailProvider returns the mail provider configured by environment variables.
// Production code uses this, while tests can call email.ProviderFromConfig directly.

// loadEmailConfigFile reads EMAIL_* style configuration values from a simple
// key=value file. Missing files return an empty configuration.

// getAdminEmails returns a slice of administrator email addresses. The
// configuration option ADMIN_EMAILS may provide a comma-separated list. When
// empty and a Queries value is supplied, the database is queried for
// administrator accounts. GetAdminEmails returns a slice of administrator
// addresses using this logic.
func GetAdminEmails(ctx context.Context, q *db.Queries) []string {
	env := config.AppRuntimeConfig.AdminEmails
	if env == "" {
		env = os.Getenv(config.EnvAdminEmails)
	}
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
		for _, e := range rows {
			if e != "" {
				emails = append(emails, e)
			}
		}
	}
	return emails
}

// AdminNotificationsEnabled reports whether administrator notification emails
// should be sent based on the runtime configuration.
func AdminNotificationsEnabled() bool {
	return config.AppRuntimeConfig.AdminNotify
}

// EmailSendingEnabled reports if queued emails should be dispatched according
// to the runtime configuration.
func EmailSendingEnabled() bool {
	return config.AppRuntimeConfig.EmailEnabled
}

// notifyThreadSubscribers emails users subscribed to the forum thread.
func NotifyThreadSubscribers(ctx context.Context, provider email.Provider, q *db.Queries, threadID, excludeUser int32, page string) {
	if q == nil {
		return
	}
	rows, err := q.ListUsersSubscribedToThread(ctx, db.ListUsersSubscribedToThreadParams{
		ForumthreadID: threadID,
		Idusers:       excludeUser,
	})
	if err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
		return
	}
	for _, row := range rows {
		if err := CreateEmailTemplateAndQueue(ctx, q, row.Idusers, row.Email, page, "update", nil); err != nil {
			log.Printf("Error: queue: %s", err)
		}
	}
}

// NotifyNewsSubscribers queues update emails for users subscribed to the given news post.
func NotifyNewsSubscribers(ctx context.Context, q *db.Queries, newsID, excludeUser int32, page string) {
	if q == nil {
		return
	}
	rows, err := q.ListUsersSubscribedToNews(ctx, db.ListUsersSubscribedToNewsParams{
		Idsitenews: newsID,
		Idusers:    excludeUser,
	})
	if err != nil {
		log.Printf("Error: listUsersSubscribedToNews: %v", err)
		return
	}
	for _, row := range rows {
		if err := CreateEmailTemplateAndQueue(ctx, q, row.Idusers, row.Email, page, "update", nil); err != nil {
			log.Printf("Error: notifyChange: %v", err)
		}
	}
}
