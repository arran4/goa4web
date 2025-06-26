package forum

import (
	"context"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	"github.com/arran4/goa4web/runtimeconfig"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr, page string) error {
	return emailutil.NotifyChange(ctx, provider, emailAddr, page)
}

func getEmailProvider() email.Provider {
	return email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig)
}

// getAdminEmails returns a slice of administrator email addresses. Environment
// variable ADMIN_EMAILS takes precedence over the database.
func getAdminEmails(ctx context.Context, q *Queries) []string {
	return emailutil.GetAdminEmails(ctx, q)
}

func adminNotificationsEnabled() bool {
	return emailutil.AdminNotificationsEnabled()
}

func notifyAdmins(ctx context.Context, provider email.Provider, q *Queries, page string) {
	emailutil.NotifyAdmins(ctx, provider, q, page)
}

func notifyThreadSubscribers(ctx context.Context, provider email.Provider, q *Queries, threadID, excludeUser int32, page string) {
	emailutil.NotifyThreadSubscribers(ctx, provider, q, threadID, excludeUser, page)
}
