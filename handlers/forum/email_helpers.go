package forum

import (
	"context"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/email"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func notifyThreadSubscribers(ctx context.Context, provider email.Provider, q *db.Queries, threadID, excludeUser int32, page string) {
	notif.Notifier{EmailProvider: provider, Queries: q}.NotifyThreadSubscribers(ctx, threadID, excludeUser, page)
}
