package notifications

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/db"
	"log"
)

func (n *Notifier) dlqRecordAndNotify(ctx context.Context, q dlq.DLQ, msg string) error {
	if q == nil {
		return fmt.Errorf("no dlq provider")
	}
	if err := q.Record(ctx, msg); err != nil {
		return err
	}
	if n.Queries == nil || !n.Config.AdminNotify {
		return nil
	}
	if dbq, ok := q.(db.DLQ); ok {
		if count, err := dbq.Queries.CountDeadLettersSystem(ctx); err == nil {
			if isPow10(count) {
				data := EmailData{
					Item: struct {
						Message string
					}{Message: msg},
				}
				et := &EmailTemplates{
					Text:    EmailTextTemplateFilenameGenerator("dlqMultiFailure"),
					HTML:    EmailHTMLTemplateFilenameGenerator("dlqMultiFailure"),
					Subject: EmailSubjectTemplateFilenameGenerator("dlqMultiFailure"),
				}
				if err := n.NotifyAdmins(ctx, et, data); err != nil {
					return err
				}
				nt, err := n.renderNotification(ctx, NotificationTemplateFilenameGenerator("dlqMultiFailure"), data)
				if err == nil {
					for _, addr := range n.adminEmails(ctx) {
						u, err := n.Queries.UserByEmail(ctx, addr)
						if err != nil {
							continue
						}
						if err := n.sendInternalNotification(ctx, u.Idusers, "", string(nt)); err != nil {
							log.Printf("send internal notification: %v", err)
						}
					}
				}
			}
		}
	}
	return nil
}

func isPow10(n int64) bool {
	if n < 1 {
		return false
	}
	for n%10 == 0 {
		n /= 10
	}
	return n == 1
}
