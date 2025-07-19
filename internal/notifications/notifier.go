package notifications

import (
	"context"
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
func NotifyAdmins(ctx context.Context, n Notifier, et *EmailTemplates, data EmailData) error {
	if !config.AdminNotificationsEnabled() {
		return nil
	}
	for _, addr := range config.GetAdminEmails(ctx, n.Queries) {
		var uid int32
		if n.Queries != nil {
			if u, err := n.Queries.UserByEmail(ctx, addr); err == nil {
				uid = u.Idusers
			} else {
				log.Printf("notify admin %s: %v", addr, err)
				continue
			}
		}
		if err := RenderAndQueueEmailFromTemplates(ctx, n.Queries, uid, addr, et, data); err != nil {
			log.Printf("notify admin %s: %v", addr, err)
		}
	}
	return nil
}
