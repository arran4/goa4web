package admin

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/mail"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/email"
)

// SaveTemplateTask stores a custom update email template.
type SaveTemplateTask struct{ tasks.TaskString }

var saveTemplateTask = &SaveTemplateTask{TaskString: TaskUpdate}

// compile-time interface check for SaveTemplateTask
var _ tasks.Task = (*SaveTemplateTask)(nil)

// TestTemplateTask queues an email using the template for preview.
type TestTemplateTask struct{ tasks.TaskString }

var testTemplateTask = &TestTemplateTask{TaskString: TaskTestMail}

// compile-time interface check for TestTemplateTask
var _ tasks.Task = (*TestTemplateTask)(nil)

func getUpdateEmailText(ctx context.Context) string {
	if cd, ok := ctx.Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
		if q := cd.Queries(); q != nil {
			if body, err := q.GetTemplateOverride(ctx, "updateEmail.gotxt"); err == nil && body != "" {
				return body
			}
		}
	}
	tmpl := templates.GetCompiledEmailTextTemplates(map[string]any{})
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "updateEmail.gotxt", nil); err != nil {
		return ""
	}
	return buf.String()
}

// AdminEmailTemplatePage allows administrators to edit the update email template.
func AdminEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
	b := getUpdateEmailText(r.Context())

	var preview string
	tmpl, err := template.New("email").Parse(b)
	if err == nil {
		var buf bytes.Buffer
		tmpl.Execute(&buf, struct{ To, From, Subject, URL string }{
			To:      "test@example.com",
			From:    config.AppRuntimeConfig.EmailFrom,
			Subject: "Website Update Notification",
			URL:     "http://example.com/page",
		})
		preview = buf.String()
	}

	data := struct {
		*common.CoreData
		Body    string
		Preview string
		Error   string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Body:     b,
		Preview:  preview,
		Error:    r.URL.Query().Get("error"),
	}

	handlers.TemplateHandler(w, r, "emailTemplatePage.gohtml", data)
}

func (SaveTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	body := r.PostFormValue("body")
	q := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := q.SetTemplateOverride(r.Context(), db.SetTemplateOverrideParams{Name: "updateEmail", Body: body}); err != nil {
		return fmt.Errorf("db save template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler("/admin/email/template")
}

func (TestTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	if email.ProviderFromConfig(config.AppRuntimeConfig) == nil {
		return fmt.Errorf("mail not configured %w", handlers.ErrRedirectOnSamePageHandler(userhandlers.ErrMailNotConfigured))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	urow, err := queries.GetUserById(r.Context(), cd.UserID)
	if err != nil {
		return fmt.Errorf("get user fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !urow.Email.Valid || urow.Email.String == "" {
		return fmt.Errorf("email unknown %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("email unknown")))
	}

	base := "http://" + r.Host
	if config.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/")
	}
	pageURL := base + r.URL.Path

	var buf bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(r.Context()))
	if err != nil {
		return fmt.Errorf("parse template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	unsub := "/usr/subscriptions"
	if config.AppRuntimeConfig.HTTPHostname != "" {
		unsub = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + unsub
	}
	content := struct{ To, From, Subject, URL, Action, Path, Time, UnsubscribeUrl string }{
		To:             (&mail.Address{Name: urow.Username.String, Address: urow.Email.String}).String(),
		From:           config.AppRuntimeConfig.EmailFrom,
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

	if f, err := mail.ParseAddress(config.AppRuntimeConfig.EmailFrom); err == nil {
		fromAddr = *f
	} else {
		fromAddr = mail.Address{Address: config.AppRuntimeConfig.EmailFrom}
	}
	msg, err := email.BuildMessage(fromAddr, toAddr, content.Subject, buf.String(), "")
	if err != nil {
		return fmt.Errorf("build message fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := queries.InsertPendingEmail(r.Context(), db.InsertPendingEmailParams{ToUserID: sql.NullInt32{Int32: urow.Idusers, Valid: true}, Body: string(msg), DirectEmail: false}); err != nil {
		return fmt.Errorf("queue email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler("/admin/email/template")
}
