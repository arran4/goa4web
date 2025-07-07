package forum

import (
	"context"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr, page string) error {
	n := notif.Notifier{EmailProvider: provider}
	return n.NotifyChange(ctx, 0, emailAddr, page, "update", nil)
}

// getAdminEmails returns a slice of administrator email addresses. Environment
// variable ADMIN_EMAILS takes precedence over the database.
func getAdminEmails(ctx context.Context, q *db.Queries) []string {
	return emailutil.GetAdminEmails(ctx, q)
}

func adminNotificationsEnabled() bool {
	return emailutil.AdminNotificationsEnabled()
}

func notifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	notif.Notifier{EmailProvider: provider, Queries: q}.NotifyAdmins(ctx, page)
}

func notifyThreadSubscribers(ctx context.Context, provider email.Provider, q *db.Queries, threadID, excludeUser int32, page string) {
	notif.Notifier{EmailProvider: provider, Queries: q}.NotifyThreadSubscribers(ctx, threadID, excludeUser, page)
}
