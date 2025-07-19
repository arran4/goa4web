package notifications

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/db"
)

func dlqRecordAndNotify(ctx context.Context, q dlq.DLQ, n Notifier, msg string) error {
	if q == nil {
		return fmt.Errorf("no dlq provider")
	}
	if err := q.Record(ctx, msg); err == nil {
		if dbq, ok := q.(db.DLQ); ok {
			if count, err := dbq.Queries.CountDeadLetters(ctx); err == nil {
				if isPow10(count) {
					// TODO create template and data
					err := NotifyAdmins(ctx, n, &EmailTemplates{
						Text:    EmailTextTemplateFilenameGenerator("dlqMultiFailure"),
						HTML:    EmailHTMLTemplateFilenameGenerator("dlqMultiFailure"),
						Subject: EmailSubjectTemplateFilenameGenerator("dlqMultiFailure"),
					}, EmailData{
						Item: struct {
							Message string
						}{
							Message: msg,
						},
					})
					if err != nil {
						return err
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
