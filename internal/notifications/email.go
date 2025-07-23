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

	"github.com/arran4/goa4web/config"
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
func (n *Notifier) renderAndQueueEmailFromTemplates(ctx context.Context, userID *int32, emailAddr string, et *EmailTemplates, data interface{}) error {
	// userID == nil implies the email is direct
	if n.Queries == nil {
		return fmt.Errorf("no query")
	}
	msg, err := n.RenderEmailFromTemplates(ctx, emailAddr, et, data)
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
}

// RenderEmailFromTemplates returns the rendered email message using the provided templates.
func (n *Notifier) RenderEmailFromTemplates(ctx context.Context, emailAddr string, et *EmailTemplates, item interface{}) ([]byte, error) {
	if emailAddr == "" {
		return nil, fmt.Errorf("no email specified")
	}
	from := email.ParseAddress(config.AppRuntimeConfig.EmailFrom)
	to := email.ParseAddress(emailAddr)

	subjectPrefix := config.AppRuntimeConfig.EmailSubjectPrefix
	if subjectPrefix == "" {
		subjectPrefix = "goa4web"
	}

	unsub := "/usr/subscriptions"
	if config.AppRuntimeConfig.HTTPHostname != "" {
		unsub = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + unsub
	}

	signOff := config.AppRuntimeConfig.EmailSignOff
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
	var textBody, htmlBody string
	subject := "[" + subjectPrefix + "] Website Update Notification"
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
	if err := eventbus.DefaultBus.Publish(evt); err != nil && err != eventbus.ErrBusClosed {
		log.Printf("publish email queue event: %v", err)
	}
	return nil
}

// sendSubscriberEmail queues an email notification for a subscriber.
func (n *Notifier) sendSubscriberEmail(ctx context.Context, userID int32, evt eventbus.TaskEvent, et *EmailTemplates) error {
	user, err := n.Queries.GetUserById(ctx, userID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		if nmErr := notifyMissingEmail(ctx, n.Queries, userID); nmErr != nil {
			log.Printf("notify missing email: %v", nmErr)
		}
		return err
	}
	if et == nil {
		return nil
	}
	return n.renderAndQueueEmailFromTemplates(ctx, &userID, user.Email.String, et, evt.Data)
}
