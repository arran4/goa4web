package admin

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	notif "github.com/arran4/goa4web/internal/notifications"
)

// TestTemplateTask queues an email using the template for preview.
type TestTemplateTask struct{ tasks.TaskString }

var testTemplateTask = &TestTemplateTask{TaskString: TaskTestMail}

// compile-time interface check for TestTemplateTask
var _ tasks.Task = (*TestTemplateTask)(nil)
var _ tasks.AuditableTask = (*TestTemplateTask)(nil)

func (TestTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.EmailProvider() == nil {
		return fmt.Errorf("mail not configured %w", handlers.ErrRedirectOnSamePageHandler(userhandlers.ErrMailNotConfigured))
	}

	queries := cd.Queries()
	urow, err := queries.SystemGetUserByID(r.Context(), cd.UserID)
	if err != nil {
		return fmt.Errorf("get user fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !urow.Email.Valid || urow.Email.String == "" {
		return fmt.Errorf("email unknown %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("email unknown")))
	}

	base := "http://" + r.Host
	if cd.Config.BaseURL != "" {
		base = strings.TrimRight(cd.Config.BaseURL, "/")
	}
	pageURL := base + r.URL.Path

	var buf bytes.Buffer
	tmpl, err := template.New("email").Funcs(cd.Funcs(r)).Parse(notif.GetUpdateEmailText(r.Context(), queries, cd.Config))
	if err != nil {
		return fmt.Errorf("parse template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	unsub := "/usr/subscriptions"
	if cd.Config.BaseURL != "" {
		unsub = strings.TrimRight(cd.Config.BaseURL, "/") + unsub
	}
	content := struct{ To, From, Subject, URL, Action, Path, Time, UnsubscribeUrl string }{
		To:             (&mail.Address{Name: urow.Username.String, Address: urow.Email.String}).String(),
		From:           cd.Config.EmailFrom,
		Subject:        "Website Update Notification",
		URL:            pageURL,
		Action:         string(TaskTestMail),
		Path:           r.URL.Path,
		Time:           time.Now().Format(time.RFC822),
		UnsubscribeUrl: unsub,
	}
	if err := tmpl.Execute(&buf, content); err != nil {
		return fmt.Errorf("execute template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	toAddr := mail.Address{Name: urow.Username.String, Address: urow.Email.String}
	var fromAddr mail.Address

	if f, err := mail.ParseAddress(cd.Config.EmailFrom); err == nil {
		fromAddr = *f
	} else {
		fromAddr = mail.Address{Address: cd.Config.EmailFrom}
	}
	msg, err := email.BuildMessage(fromAddr, toAddr, content.Subject, buf.String(), "")
	if err != nil {
		return fmt.Errorf("build message fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := queries.InsertPendingEmail(r.Context(), db.InsertPendingEmailParams{ToUserID: sql.NullInt32{Int32: urow.Idusers, Valid: true}, Body: string(msg), DirectEmail: false}); err != nil {
		return fmt.Errorf("queue email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd != nil {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Template"] = "updateEmail"
			evt.Data["PreviewEmail"] = urow.Email.String
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/email/template"}
}

// AuditRecord summarises sending a preview email.
func (TestTemplateTask) AuditRecord(data map[string]any) string {
	if addr, ok := data["PreviewEmail"].(string); ok {
		return "sent preview email to " + addr
	}
	return "sent preview email"
}
