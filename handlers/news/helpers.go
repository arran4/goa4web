package news

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

func processCommentFullQuote(username, text string) string {
	var out bytes.Buffer
	var quote bytes.Buffer
	var it, bc, nlc int

	for it < len(text) {
		switch text[it] {
		case ']':
			bc--
		case '[':
			bc++
		case '\\':
			if it+1 < len(text) {
				if text[it+1] == '[' || text[it+1] == ']' {
					out.WriteByte(text[it+1])
					it++
				}
			}
		case '\n':
			if bc == 0 && nlc == 1 {
				quote.WriteString(processCommentQuote(username, out.String()))
				out.Reset()
			}
			nlc++
			it++
			continue
		case '\r':
			it++
			continue
		case ' ':
			fallthrough
		default:
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			out.WriteByte(text[it])
		}
		it++
	}
	quote.WriteString(processCommentQuote(username, out.String()))
	return quote.String()
}

func processCommentQuote(username string, text string) string {
	return fmt.Sprintf("[quoteof \"%s\" %s]\n", username, text)
}

func notifyChange(ctx context.Context, provider email.Provider, emailAddr string, page string) error {
	if emailAddr == "" {
		return fmt.Errorf("no email specified")
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

	var notification bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}
	if err := tmpl.Execute(&notification, content); err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}

	if q, ok := ctx.Value(hcommon.KeyQueries).(*db.Queries); ok {
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

func PostUpdateLocal(ctx context.Context, q *db.Queries, threadID, topicID int32) error {
	if err := q.RecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.RebuildForumTopicByIdMetaColumns(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}

func getUpdateEmailText(ctx context.Context) string {
	// simplified; in original package it reads from templates. Here we just return a basic text.
	return "A page you follow has been updated: {{.URL}}"
}
