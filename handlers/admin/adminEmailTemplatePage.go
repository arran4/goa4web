package admin

import (
	"bytes"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"text/template"
	"time"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	handlers "github.com/arran4/goa4web/handlers"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	emailutil "github.com/arran4/goa4web/internal/notifications"
)

type saveTemplateTask struct{ tasks.TaskString }
type testTemplateTask struct{ tasks.TaskString }

// AdminEmailTemplatePage allows administrators to edit the update email template.
func AdminEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
	b := emailutil.GetUpdateEmailText(r.Context())

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
		*CoreData
		Body    string
		Preview string
		Error   string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Body:     b,
		Preview:  preview,
		Error:    r.URL.Query().Get("error"),
	}

	handlers.TemplateHandler(w, r, "emailTemplatePage.gohtml", data)
}

func (saveTemplateTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body := r.PostFormValue("body")
	q := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := q.SetTemplateOverride(r.Context(), db.SetTemplateOverrideParams{Name: "updateEmail", Body: body}); err != nil {
		log.Printf("db save template: %v", err)
	}
	http.Redirect(w, r, "/admin/email/template", http.StatusSeeOther)
}

func (testTemplateTask) Action(w http.ResponseWriter, r *http.Request) {
	if email.ProviderFromConfig(config.AppRuntimeConfig) == nil {
		q := url.QueryEscape(userhandlers.ErrMailNotConfigured.Error())
		r.URL.RawQuery = "error=" + q
		handlers.TaskErrorAcknowledgementPage(w, r)
		return
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	urow, err := queries.GetUserById(r.Context(), cd.UserID)
	if err != nil {
		log.Printf("get user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !urow.Email.Valid || urow.Email.String == "" {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}

	base := "http://" + r.Host
	if config.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/")
	}
	pageURL := base + r.URL.Path

	var buf bytes.Buffer
	tmpl, err := template.New("email").Parse(emailutil.GetUpdateEmailText(r.Context()))
	if err != nil {
		log.Printf("parse template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	unsub := "/usr/subscriptions"
	if config.AppRuntimeConfig.HTTPHostname != "" {
		unsub = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + unsub
	}
	content := struct{ To, From, Subject, URL, Action, Path, Time, UnsubURL string }{
		To:       (&mail.Address{Name: urow.Username.String, Address: urow.Email.String}).String(),
		From:     config.AppRuntimeConfig.EmailFrom,
		Subject:  "Website Update Notification",
		URL:      pageURL,
		Action:   TaskTestMail,
		Path:     r.URL.Path,
		Time:     time.Now().Format(time.RFC822),
		UnsubURL: unsub,
	}
	if err := tmpl.Execute(&buf, content); err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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
		log.Printf("build message: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.InsertPendingEmail(r.Context(), db.InsertPendingEmailParams{ToUserID: urow.Idusers, Body: string(msg)}); err != nil {
		log.Printf("queue email: %v", err)
	}
	http.Redirect(w, r, "/admin/email/template", http.StatusSeeOther)
}
