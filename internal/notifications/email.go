package notifications

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"

	"github.com/arran4/goa4web/config"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

func getEmailTemplates(ctx context.Context, action string) (string, string) {
	// Compile embedded templates so overrides work.
	_ = templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	_ = templates.GetCompiledEmailTextTemplates(map[string]any{})

	name := "email_" + strings.ToLower(action)
	nameHTML := name + "_html"
	var text, html string
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, name); err == nil && body != "" {
			text = body
		}
		if body, err := q.GetTemplateOverride(ctx, nameHTML); err == nil && body != "" {
			html = body
		}
	}
	return text, html
}

// TODO: consider making this private and replacing with EmailTemplates.CreateEmail.
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

// TODO: make private once call sites are updated.
func CreateEmailTemplateAndQueue(ctx context.Context, q *db.Queries, userID int32, emailAddr, page, action string, item interface{}) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	msg, _, err := CreateEmailTemplate(ctx, emailAddr, page, action, item)
	if err != nil {
		return err
	}
	return queueEmail(ctx, q, userID, msg)
}

// QueueEmailFromTemplates renders the provided templates and queues the result.
// TODO: make private and unify call sites.
func QueueEmailFromTemplates(ctx context.Context, q *db.Queries, userID int32, emailAddr string, et *EmailTemplates, data interface{}) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	msg, err := RenderEmailFromTemplates(emailAddr, et, data)
	if err != nil {
		return err
	}
	return queueEmail(ctx, q, userID, msg)
}

// RenderEmailFromTemplates returns the rendered email message using the provided templates.
// TODO: evaluate exposing this via EmailTemplates.CreateEmail instead.
func RenderEmailFromTemplates(emailAddr string, et *EmailTemplates, data interface{}) ([]byte, error) {
	if et == nil || emailAddr == "" {
		return nil, fmt.Errorf("invalid args")
	}
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})

	prefix := config.AppRuntimeConfig.EmailSubjectPrefix
	if prefix == "" {
		prefix = "goa4web"
	}

	content := struct {
		To            string
		From          string
		SubjectPrefix string
		Item          interface{}
	}{
		To:            emailAddr,
		From:          config.AppRuntimeConfig.EmailFrom,
		SubjectPrefix: prefix,
		Item:          data,
	}

	var textBody, htmlBody, subject string
	if et.Text != "" {
		var buf bytes.Buffer
		if err := textTmpls.ExecuteTemplate(&buf, et.Text, content); err != nil {
			return nil, err
		}
		textBody = buf.String()
	}
	if et.HTML != "" {
		var buf bytes.Buffer
		if err := htmlTmpls.ExecuteTemplate(&buf, et.HTML, content); err != nil {
			return nil, err
		}
		htmlBody = buf.String()
	}
	if et.Subject != "" {
		var buf bytes.Buffer
		if err := textTmpls.ExecuteTemplate(&buf, et.Subject, content); err != nil {
			return nil, err
		}
		subject = strings.TrimSpace(buf.String())
	}
	from := email.ParseAddress(config.AppRuntimeConfig.EmailFrom)
	to := email.ParseAddress(emailAddr)
	return email.BuildMessage(from, to, subject, textBody, htmlBody)
}

func queueEmail(ctx context.Context, q *db.Queries, userID int32, msg []byte) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	return q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: userID, Body: string(msg)})
}
