package admin

import (
	"context"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	notif "github.com/arran4/goa4web/internal/notifications"
)

// notifyAdmins sends a change notification to administrator addresses.
func notifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	notif.Notifier{EmailProvider: provider, Queries: q}.NotifyAdmins(ctx, page)
}
