package notifications

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/internal/eventbus"
	"strings"

	"github.com/arran4/goa4web/config"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

// TODO: make private once call sites are updated.
func (n *Notifier) CreateEmailTemplateAndQueue(ctx context.Context, userID int32, emailAddr, page, action string, item interface{}) error {
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
	return n.queueEmail(ctx, userID, msg)
}

// RenderAndQueueEmailFromTemplates renders the provided templates and queues the result.
// TODO: make private and unify call sites.
func (n *Notifier) RenderAndQueueEmailFromTemplates(ctx context.Context, userID int32, emailAddr string, et *EmailTemplates, data interface{}) error {
	if n.Queries == nil {
		return fmt.Errorf("no query")
	}
	msg, err := n.RenderEmailFromTemplates(ctx, emailAddr, et, data)
	if err != nil {
		return err
	}
	return n.queueEmail(ctx, userID, msg)
}

type EmailData struct {
	any
	SubjectPrefix  string
	UnsubscribeUrl string
	Item           interface{}
}

// RenderEmailFromTemplates returns the rendered email message using the provided templates.
// TODO: evaluate exposing this via EmailTemplates.CreateEmail instead.
// TODO: this should be a receiver on EmailTemplates
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

	data := EmailData{
		any:            item,
		SubjectPrefix:  subjectPrefix,
		UnsubscribeUrl: unsub,
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
		tb, err := n.renderEmailText(ctx, et.Subject, data)
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

func (n *Notifier) queueEmail(ctx context.Context, userID int32, msg []byte) error {
	if n.Queries == nil {
		return fmt.Errorf("no query")
	}
	return n.Queries.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: userID, Body: string(msg)})
}

// sendSubscriberEmail queues an email notification for a subscriber.
func (n *Notifier) sendSubscriberEmail(ctx context.Context, userID int32, evt eventbus.Event, et *EmailTemplates) error {
	user, err := n.Queries.GetUserById(ctx, userID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		notifyMissingEmail(ctx, n.Queries, userID)
		return err
	}
	if et == nil {
		return nil
	}
	return n.RenderAndQueueEmailFromTemplates(ctx, userID, user.Email.String, et, evt.Data)
}
