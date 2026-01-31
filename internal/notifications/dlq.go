package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

func (n *Notifier) dlqRecordAndNotify(ctx context.Context, q dlq.DLQ, msg string, evt *eventbus.TaskEvent) error {
	if q == nil {
		return fmt.Errorf("no dlq provider")
	}
	dlqMsg := dlq.Message{
		Error: msg,
		Event: evt,
	}
	if evt != nil && evt.Task != nil {
		if t, ok := evt.Task.(tasks.Name); ok {
			dlqMsg.TaskName = t.Name()
		} else if ts, ok := evt.Task.(tasks.TaskString); ok {
			dlqMsg.TaskName = string(ts)
		}
	}
	recordMsg := msg
	if b, err := json.Marshal(dlqMsg); err == nil {
		recordMsg = string(b)
	}

	if err := q.Record(ctx, recordMsg); err != nil {
		return err
	}
	if n.Queries == nil || !n.Config.AdminNotify {
		return nil
	}
	if dbq, ok := q.(db.DLQ); ok {
		if count, err := dbq.Queries.SystemCountDeadLetters(ctx); err == nil {
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
						u, err := n.Queries.SystemGetUserByEmail(ctx, addr)
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
