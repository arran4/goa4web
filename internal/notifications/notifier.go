package notifications

import (
	"context"
	"database/sql"
	"log"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

// Notifier dispatches updates via email and internal notifications.
// Notifier dispatches updates via email and internal notifications.
type Notifier struct {
	EmailProvider email.Provider
	Queries       *dbpkg.Queries
}

// NotifyAdmins sends a generic update notice to administrator accounts.
func NotifyAdmins(ctx context.Context, n Notifier, page string) {
	if !config.AdminNotificationsEnabled() {
		return
	}
	for _, addr := range config.GetAdminEmails(ctx, n.Queries) {
		var uid int32
		if n.Queries != nil {
			if u, err := n.Queries.UserByEmail(ctx, sql.NullString{String: addr, Valid: true}); err == nil {
				uid = u.Idusers
			} else {
				log.Printf("user by email %s: %v", addr, err)
			}
		}
		if err := CreateEmailTemplateAndQueue(ctx, n.Queries, uid, addr, page, "update", nil); err != nil {
			log.Printf("notify admin %s: %v", addr, err)
		}
	}
}
