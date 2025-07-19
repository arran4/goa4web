package notifications

import (
	"context"
	"log"
	"sync"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	htemplate "html/template"
)

// Notifier dispatches updates via email and internal notifications.
// Notifier dispatches updates via email and internal notifications.
type Notifier struct {
	EmailProvider  email.Provider
	Queries        *dbpkg.Queries
	noteOnce       sync.Once
	noteTmpls      *htemplate.Template
	emailTextOnce  sync.Once
	emailTextTmpls *htemplate.Template
	emailHTMLOnce  sync.Once
	emailHTMLTmpls *htemplate.Template
}

// New constructs a Notifier with the provided dependencies.
func New(q *dbpkg.Queries, p email.Provider) *Notifier {
	return &Notifier{Queries: q, EmailProvider: p}
}

func (n *Notifier) notificationTemplates() *htemplate.Template {
	n.noteOnce.Do(func() {
		n.noteTmpls = templates.GetCompiledNotificationTemplates(map[string]any{})
	})
	return n.noteTmpls
}

func (n *Notifier) emailTextTemplates() *htemplate.Template {
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
		if err := n.RenderAndQueueEmailFromTemplates(ctx, uid, addr, et, data); err != nil {
			log.Printf("notify admin %s: %v", addr, err)
		}
	}
	return nil
}
