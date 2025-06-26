package forum

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/arran4/goa4web/config"
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
	env := os.Getenv(config.EnvAdminEmails)
	var emails []string
	if env != "" {
		for _, e := range strings.Split(env, ",") {
			if addr := strings.TrimSpace(e); addr != "" {
				emails = append(emails, addr)
			}
		}
		return emails
	}
	if q != nil {
		rows, err := q.ListAdministratorEmails(ctx)
		if err != nil {
			log.Printf("list admin emails: %v", err)
			return emails
		}
		for _, email := range rows {
			if email.Valid {
				emails = append(emails, email.String)
			}
		}
	}
	return emails
}

func adminNotificationsEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvAdminNotify))
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

func notifyAdmins(ctx context.Context, provider email.Provider, q *Queries, page string) {
	if provider == nil || !adminNotificationsEnabled() {
		return
	}
	for _, addr := range getAdminEmails(ctx, q) {
		if err := notifyChange(ctx, provider, addr, page); err != nil {
			log.Printf("Error: notifyChange: %s", err)
		}
	}
}

func notifyThreadSubscribers(ctx context.Context, provider email.Provider, q *Queries, threadID, excludeUser int32, page string) {
	emailutil.NotifyThreadSubscribers(ctx, provider, q, threadID, excludeUser, page)
}
