package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	htemplate "html/template"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
)

func (n *Notifier) createEmailTemplateAndQueue(ctx context.Context, userID *int32, emailAddr, page, action string, item interface{}) error {
	if n.Queries == nil {
		return fmt.Errorf("no query")
	}
	prefix := strings.ToLower(action) + "Email"
	et := NewEmailTemplates(prefix)
	data := map[string]any{"page": page, "item": item}
	msg, err := n.RenderEmailFromTemplates(ctx, emailAddr, et, data)
	if err != nil {
		return err
	}
	return n.queueEmail(ctx, userID, false, msg)
}

// renderAndQueueEmailFromTemplates renders the provided templates and queues the result.
func (n *Notifier) renderAndQueueEmailFromTemplates(ctx context.Context, userID *int32, emailAddr string, et *EmailTemplates, data interface{}, opts ...EmailOption) error {
	// userID == nil implies the email is direct
	if n.Queries == nil {
		return fmt.Errorf("no query")
	}
	msg, err := n.RenderEmailFromTemplates(ctx, emailAddr, et, data, opts...)
	if err != nil {
		return err
	}
	direct := userID == nil
	return n.queueEmail(ctx, userID, direct, msg)
}

type EmailData struct {
	any
	URL            string
	SubjectPrefix  string
	UnsubscribeUrl string
	SignOff        string
	SignOffHTML    htemplate.HTML
	Item           interface{}
	Recipient      *db.SystemGetUserByIDRow
}

// EmailOption configures EmailData prior to rendering.
type EmailOption func(*EmailData)

// WithAdmin appends " Admin" to the subject prefix to flag administrative emails.
func WithAdmin() EmailOption {
	return func(d *EmailData) { d.SubjectPrefix += " Admin" }
}

// WithRecipient adds the recipient user to EmailData.
func WithRecipient(u *db.SystemGetUserByIDRow) EmailOption {
	return func(d *EmailData) { d.Recipient = u }
}

// RenderEmailFromTemplates returns the rendered email message using the provided templates.
// Options may adjust the email metadata prior to rendering.
func (n *Notifier) RenderEmailFromTemplates(ctx context.Context, emailAddr string, et *EmailTemplates, item interface{}, opts ...EmailOption) ([]byte, error) {
	if emailAddr == "" {
		return nil, fmt.Errorf("no email specified")
	}
	from := email.ParseAddress(n.Config.EmailFrom)
	to := email.ParseAddress(emailAddr)

	subjectPrefix := n.Config.EmailSubjectPrefix
	if subjectPrefix == "" {
		subjectPrefix = "goa4web"
	}

	unsub := "/usr/subscriptions"
	if n.Config.HTTPHostname != "" {
		unsub = strings.TrimRight(n.Config.HTTPHostname, "/") + unsub
	}

	signOff := n.Config.EmailSignOff
	htmlSignOff := html.EscapeString(signOff)
	htmlSignOff = strings.ReplaceAll(htmlSignOff, "\n", "<br />")

	var urlStr string
	if m, ok := item.(map[string]any); ok {
		if v, ok := m["URL"]; ok {
			if s, ok := v.(string); ok {
				urlStr = s
			}
		} else if v, ok := m["page"]; ok {
			if s, ok := v.(string); ok {
				urlStr = s
			}
		}
	}

	data := EmailData{
		any:            item,
		URL:            urlStr,
		SubjectPrefix:  subjectPrefix,
		UnsubscribeUrl: unsub,
		SignOff:        signOff,
		SignOffHTML:    htemplate.HTML(htmlSignOff),
		Item:           item,
	}

	for _, opt := range opts {
		opt(&data)
	}

	var textBody, htmlBody string
	subject := "[" + data.SubjectPrefix + "] Website Update Notification"
	if et.Subject != "" {
		bs, err := n.renderEmailSubject(ctx, et.Subject, data)
		if err != nil {
			return nil, err
		}
		subject = strings.TrimSpace(string(bs))
	}
	if et.Text != "" {
		tb, err := n.renderEmailText(ctx, et.Text, data)
		if err != nil {
			return nil, err
		}
		textBody = strings.TrimSpace(string(tb))
	}
	if et.HTML != "" {
		hb, err := n.renderEmailHtml(ctx, et.HTML, data)
		if err != nil {
			return nil, err
		}
		htmlBody = strings.TrimSpace(string(hb))
	}
	return email.BuildMessage(from, to, subject, textBody, htmlBody)
}

func (n *Notifier) queueEmail(ctx context.Context, userID *int32, direct bool, msg []byte) error {
	if n.Queries == nil {
		return fmt.Errorf("no query")
	}
	var uid sql.NullInt32
	if userID != nil {
		uid = sql.NullInt32{Int32: *userID, Valid: !direct}
	}
	if err := n.Queries.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: uid, Body: string(msg), DirectEmail: direct}); err != nil {
		return err
	}
	evt := eventbus.EmailQueueEvent{Time: time.Now()}
	if n.Bus != nil {
		if err := n.Bus.Publish(evt); err != nil && err != eventbus.ErrBusClosed {
			log.Printf("publish email queue event: %v", err)
		}
	}
	return nil
}

// sendSubscriberEmail queues an email notification for a subscriber.
func (n *Notifier) sendSubscriberEmail(ctx context.Context, userID int32, evt eventbus.TaskEvent, et *EmailTemplates) error {
	user, err := n.Queries.SystemGetUserByID(ctx, userID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		if nmErr := notifyMissingEmail(ctx, n.Queries, userID); nmErr != nil {
			log.Printf("notify missing email: %v", nmErr)
		}
		return err
	}
	if et == nil {
		return nil
	}
	return n.renderAndQueueEmailFromTemplates(ctx, &userID, user.Email.String, et, evt.Data, WithRecipient(user))
}
