package notifications

import (
	"context"
	"database/sql"
	"log"
	"sync"
	ttemplate "text/template"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	htemplate "html/template"

	dbstart "github.com/arran4/goa4web/internal/app/dbstart"
)

// Notifier dispatches updates via email and internal notifications.
// Notifier dispatches updates via email and internal notifications.
type Notifier struct {
	EmailProvider  email.Provider
	Queries        *dbpkg.Queries
	noteOnce       sync.Once
	noteTmpls      *ttemplate.Template
	emailTextOnce  sync.Once
	emailTextTmpls *ttemplate.Template
	emailHTMLOnce  sync.Once
	emailHTMLTmpls *htemplate.Template
}

// Option configures a Notifier instance.
type Option func(*Notifier)

// WithQueries sets the db.Queries dependency.
func WithQueries(q *dbpkg.Queries) Option { return func(n *Notifier) { n.Queries = q } }

// WithEmailProvider sets the email provider dependency.
func WithEmailProvider(p email.Provider) Option { return func(n *Notifier) { n.EmailProvider = p } }

// WithConfig derives dependencies from cfg when they are not supplied.
func WithConfig(cfg config.RuntimeConfig) Option {
	return func(n *Notifier) {
		if n.EmailProvider == nil {
			n.EmailProvider = email.ProviderFromConfig(cfg)
		}
		if n.Queries == nil {
			if db := dbstart.GetDBPool(); db != nil {
				n.Queries = dbpkg.New(db)
			}
		}
	}
}

// New constructs a Notifier with the provided dependencies.
func New(opts ...Option) *Notifier {
	n := &Notifier{}
	for _, o := range opts {
		o(n)
	}
	WithConfig(config.AppRuntimeConfig)(n)
	return n
}

func (n *Notifier) notificationTemplates() *ttemplate.Template {
	n.noteOnce.Do(func() {
		n.noteTmpls = templates.GetCompiledNotificationTemplates(map[string]any{})
	})
	return n.noteTmpls
}

func (n *Notifier) emailTextTemplates() *ttemplate.Template {
	n.emailTextOnce.Do(func() {
		n.emailTextTmpls = templates.GetCompiledEmailTextTemplates(map[string]any{})
	})
	return n.emailTextTmpls
}

func (n *Notifier) emailHTMLTemplates() *htemplate.Template {
	n.emailHTMLOnce.Do(func() {
		n.emailHTMLTmpls = templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	})
	return n.emailHTMLTmpls
}

// NotifyAdmins sends a generic update notice to administrator accounts.
func (n *Notifier) NotifyAdmins(ctx context.Context, et *EmailTemplates, data EmailData) error {
	return n.notifyAdmins(ctx, et, nil, data, "")
}

func (n *Notifier) notifyAdmins(ctx context.Context, et *EmailTemplates, nt *string, data interface{}, link string) error {
	if n.Queries == nil {
		return nil
	}
	if !config.AdminNotificationsEnabled() {
		return nil
	}
	for _, addr := range config.GetAdminEmails(ctx, n.Queries) {
		var uid *int32
		if u, err := n.Queries.UserByEmail(ctx, addr); err == nil {
			id := u.Idusers
			uid = &id
		} else {
			log.Printf("notify admin %s: %v", addr, err)
		}
		if et != nil {
			if err := n.renderAndQueueEmailFromTemplates(ctx, uid, addr, et, data, false); err != nil {
				return err
			}
		}
		if nt != nil {
			msg, err := n.renderNotification(ctx, *nt, data)
			if err != nil {
				return err
			}
			if uid != nil {
				if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
					UsersIdusers: *uid,
					Link:         sql.NullString{String: link, Valid: link != ""},
					Message:      sql.NullString{String: string(msg), Valid: len(msg) > 0},
				}); err != nil {
					return err
				}
			} else {
				log.Printf("Error uid not found for %s in admin email template notification", addr)
			}
		}
	}
	return nil
}
