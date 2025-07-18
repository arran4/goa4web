package notifications

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
	"strings"

	"github.com/arran4/goa4web/config"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

// TODO: make private once call sites are updated.
func CreateEmailTemplateAndQueue(ctx context.Context, q *db.Queries, userID int32, emailAddr, page, action string, item interface{}) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	msg, _, err := CreateEmailFromTemplates(ctx, emailAddr, page, action, item)
	if err != nil {
		return err
	}
	return queueEmail(ctx, q, userID, msg)
}

// RenderAndQueueEmailFromTemplates renders the provided templates and queues the result.
// TODO: make private and unify call sites.
func RenderAndQueueEmailFromTemplates(ctx context.Context, q *db.Queries, userID int32, emailAddr string, et *EmailTemplates, data interface{}) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	msg, err := RenderEmailFromTemplates(ctx, q, emailAddr, et, data)
	if err != nil {
		return err
	}
	return queueEmail(ctx, q, userID, msg)
}

type EmailData struct {
	any
	SubjectPrefix  string
	UnsubscribeUrl string
	Item           interface{}
}

// RenderEmailFromTemplates returns the rendered email message using the provided templates.
// TODO: evaluate exposing this via EmailTemplates.CreateEmail instead.
func RenderEmailFromTemplates(ctx context.Context, q *db.Queries, emailAddr string, et *EmailTemplates, item interface{}) ([]byte, error) {
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
		SubjectPrefix:  subjectPrefix,
		UnsubscribeUrl: unsub,
		Item:           item,
	}
	var textBody, htmlBody string
	subject := "[" + subjectPrefix + "] Website Update Notification"
	if et.Subject != "" {
		bs, err := renderEmailSubject(ctx, q, et.Subject, data)
		if err != nil {
			return nil, err
		}
		subject = strings.TrimSpace(string(bs))
	}
	if et.Text != "" {
		tb, err := renderEmailText(ctx, q, et.Subject, data)
		if err != nil {
			return nil, err
		}
		textBody = strings.TrimSpace(string(tb))
	}
	if et.HTML != "" {
		hb, err := renderEmailHtml(ctx, q, et.HTML, data)
		if err != nil {
			return nil, err
		}
		htmlBody = strings.TrimSpace(string(hb))
	}
	return email.BuildMessage(from, to, subject, textBody, htmlBody)
}

func queueEmail(ctx context.Context, q *db.Queries, userID int32, msg []byte) error {
	if q == nil {
		return fmt.Errorf("no query")
	}
	return q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: userID, Body: string(msg)})
}

// sendSubscriberEmail queues an email notification for a subscriber.
func sendSubscriberEmail(ctx context.Context, n Notifier, userID int32, evt eventbus.Event, et *EmailTemplates) error {
	user, err := n.Queries.GetUserById(ctx, userID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		notifyMissingEmail(ctx, n.Queries, userID)
		return err
	}
	if et == nil {
		return nil
	}
	return RenderAndQueueEmailFromTemplates(ctx, n.Queries, userID, user.Email.String, et, evt.Data)
}
