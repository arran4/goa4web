package notifications

import (
	"context"
	"database/sql"
	htemplate "html/template"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	ttemplate "text/template"
	"time"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
)

// Notifier dispatches updates via email and internal notifications.
// Notifier dispatches updates via email and internal notifications.
type Notifier struct {
	Bus            *eventbus.Bus
	EmailProvider  email.Provider
	Queries        db.Querier
	Config         *config.RuntimeConfig
	noteOnce       sync.Once
	noteTmpls      *ttemplate.Template
	emailTextOnce  sync.Once
	emailTextTmpls *ttemplate.Template
	emailHTMLOnce  sync.Once
	emailHTMLTmpls *htemplate.Template
	CustomQueries  db.CustomQueries
}

// Option configures a Notifier instance.
type Option func(*Notifier)

// WithQueries sets the db.Queries dependency.
func WithQueries(q db.Querier) Option { return func(n *Notifier) { n.Queries = q } }

// WithCustomQueries sets the db.CustomQueries dependency.
func WithCustomQueries(cq db.CustomQueries) Option {
	return func(n *Notifier) { n.CustomQueries = cq }
}

// WithEmailProvider sets the email provider dependency.
func WithEmailProvider(p email.Provider) Option { return func(n *Notifier) { n.EmailProvider = p } }

// WithBus sets the event bus dependency used to publish email queue events.
func WithBus(b *eventbus.Bus) Option { return func(n *Notifier) { n.Bus = b } }

// WithConfig derives dependencies from cfg when they are not supplied.
func WithConfig(cfg *config.RuntimeConfig) Option {
	return func(n *Notifier) {
		n.Config = cfg
	}
}

// New constructs a Notifier with the provided dependencies.
func New(opts ...Option) *Notifier {
	n := &Notifier{}
	for _, o := range opts {
		o(n)
	}
	return n
}

func (n *Notifier) notificationTemplates() *ttemplate.Template {
	n.noteOnce.Do(func() {
		n.noteTmpls = templates.GetCompiledNotificationTemplates(defaultFuncs(), templates.WithDir(n.Config.TemplatesDir))
	})
	return n.noteTmpls
}

func defaultFuncs() map[string]any {
	return map[string]any{
		"a4code2string": func(s string) string {
			c := a4code2html.New()
			c.CodeType = a4code2html.CTWordsOnly
			c.SetInput(s)
			out, _ := io.ReadAll(c.Process())
			return string(out)
		},
		"truncateWords": func(i int, s string) string {
			words := strings.Fields(s)
			if len(words) > i {
				return strings.Join(words[:i], " ") + "..."
			}
			return s
		},
	}
}

func (n *Notifier) emailTextTemplates() *ttemplate.Template {
	n.emailTextOnce.Do(func() {
		n.emailTextTmpls = templates.GetCompiledEmailTextTemplates(map[string]any{}, templates.WithDir(n.Config.TemplatesDir))
	})
	return n.emailTextTmpls
}

func (n *Notifier) emailHTMLTemplates() *htemplate.Template {
	n.emailHTMLOnce.Do(func() {
		n.emailHTMLTmpls = templates.GetCompiledEmailHtmlTemplates(map[string]any{}, templates.WithDir(n.Config.TemplatesDir))
	})
	return n.emailHTMLTmpls
}

func (n *Notifier) adminEmails(ctx context.Context) []string {
	var env string
	if n.Config != nil {
		env = n.Config.AdminEmails
	}
	if env == "" {
		env = os.Getenv(config.EnvAdminEmails)
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
	if n.Queries != nil {
		rows, err := n.Queries.AdminListAdministratorEmails(ctx)
		if err != nil {
			log.Printf("AdminListAdministratorEmails: %v", err)
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

// NotifyAdmins sends a generic update notice to administrator accounts.
func (n *Notifier) NotifyAdmins(ctx context.Context, et *EmailTemplates, data EmailData) error {
	return n.notifyAdmins(ctx, et, nil, nil, data, "")
}

func (n *Notifier) notifyAdmins(ctx context.Context, et *EmailTemplates, nt *string, excludeUserID *int32, data interface{}, link string) error {
	if n.Queries == nil {
		return nil
	}
	if n.Config != nil && !n.Config.AdminNotify {
		return nil
	}
	for _, addr := range n.adminEmails(ctx) {
		var uid *int32
		if u, err := n.Queries.SystemGetUserByEmail(ctx, addr); err == nil {
			id := u.Idusers
			if excludeUserID != nil && id == *excludeUserID {
				continue
			}
			uid = &id
		} else {
			log.Printf("notify admin %s: %v", addr, err)
		}
		if et != nil {
			if err := n.renderAndQueueEmailFromTemplates(ctx, uid, addr, et, data, WithAdmin()); err != nil {
				return err
			}
		}
		if nt != nil {
			// Internal notifications expect data wrapped in .Item and .Path
			renderData := struct {
				Path string
				Item any
			}{
				Path: link,
				Item: data,
			}
			msg, err := n.renderNotification(ctx, *nt, renderData)
			if err != nil {
				return err
			}
			if uid != nil {
				if err := n.Queries.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
					RecipientID: *uid,
					Link:        sql.NullString{String: link, Valid: link != ""},
					Message:     sql.NullString{String: string(msg), Valid: len(msg) > 0},
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

// NotificationPurgeWorker periodically removes old read notifications.
func (n *Notifier) NotificationPurgeWorker(ctx context.Context, interval time.Duration) {
	if n.Queries == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := n.Queries.AdminPurgeReadNotifications(ctx); err != nil {
				log.Printf("purge notifications: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// sendInternalNotification stores an internal notification for the user.
func (n *Notifier) sendInternalNotification(ctx context.Context, userID int32, path, msg string) error {
	return n.Queries.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
		RecipientID: userID,
		Link:        sql.NullString{String: path, Valid: path != ""},
		Message:     sql.NullString{String: msg, Valid: msg != ""},
	})
}
