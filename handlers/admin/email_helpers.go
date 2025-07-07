package admin

import (
	"context"
	db "github.com/arran4/goa4web/internal/db"
	"os"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr, page string) error {
	n := notif.Notifier{EmailProvider: provider}
	return n.NotifyChange(ctx, 0, emailAddr, page, "update", nil)
}

func emailSendingEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvEmailEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

func getUpdateEmailText(ctx context.Context) string {
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, "updateEmail"); err == nil && body != "" {
			return body
		}
	}
	return defaultUpdateEmailText
}

var defaultUpdateEmailText = templates.UpdateEmailText

func getAdminEmails(ctx context.Context, q *db.Queries) []string {
	return emailutil.GetAdminEmails(ctx, q)
}

func adminNotificationsEnabled() bool {
	return emailutil.AdminNotificationsEnabled()
}

func notifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	notif.Notifier{EmailProvider: provider, Queries: q}.NotifyAdmins(ctx, page)
}
