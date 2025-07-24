package config

import (
	"context"
	"github.com/arran4/goa4web/internal/db"
	"log"
	"os"
	"strings"
)

// getEmailProvider returns the mail provider configured by environment variables.
// Production code uses this, while tests can call email.ProviderFromConfig directly.

// loadEmailConfigFile reads EMAIL_* style configuration values from a simple
// key=value file. Missing files return an empty configuration.

// getAdminEmails returns a slice of administrator email addresses. The
// configuration option ADMIN_EMAILS may provide a comma-separated list. When
// empty and a Queries value is supplied, the database is queried for
// administrator accounts. GetAdminEmails returns a slice of administrator
// addresses using this logic.
func GetAdminEmails(ctx context.Context, q *db.Queries, cfg RuntimeConfig) []string {
	env := cfg.AdminEmails
	if env == "" {
		env = os.Getenv(EnvAdminEmails)
	}
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
		for _, e := range rows {
			if e != "" {
				emails = append(emails, e)
			}
		}
	}
	return emails
}

// AdminNotificationsEnabled reports whether administrator notification emails
// should be sent based on the runtime configuration.
func AdminNotificationsEnabled(cfg RuntimeConfig) bool {
	return cfg.AdminNotify
}

// EmailSendingEnabled reports if queued emails should be dispatched according
// to the runtime configuration.
func EmailSendingEnabled(cfg RuntimeConfig) bool {
	return cfg.EmailEnabled
}
