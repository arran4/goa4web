package admin

import (
	"bytes"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// AdminEmailTemplatePage allows administrators to edit the update email template.
func AdminEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
	b := getUpdateEmailText(r.Context())

	var preview string
	tmpl, err := template.New("email").Parse(b)
	if err == nil {
		var buf bytes.Buffer
		tmpl.Execute(&buf, struct{ To, From, Subject, URL string }{
			To:      "test@example.com",
			From:    email.SourceEmail,
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

	if err := templates.RenderTemplate(w, "emailTemplatePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminEmailTemplateSaveActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body := r.PostFormValue("body")
	q := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := q.SetTemplateOverride(r.Context(), "updateEmail", body); err != nil {
		log.Printf("db save template: %v", err)
	}
	http.Redirect(w, r, "/admin/email/template", http.StatusSeeOther)
}

func AdminEmailTemplateTestActionPage(w http.ResponseWriter, r *http.Request) {
	provider := getEmailProvider()
	if provider == nil {
		q := url.QueryEscape(userhandlers.ErrMailNotConfigured)
		http.Redirect(w, r, "/admin/email/template?error="+q, http.StatusTemporaryRedirect)
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
	user := &db.User{Idusers: urow.Idusers, Email: urow.Email, Passwd: urow.Passwd, PasswdAlgorithm: urow.PasswdAlgorithm, Username: urow.Username}
	if !user.Email.Valid {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}

	base := "http://" + r.Host
	if runtimeconfig.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(runtimeconfig.AppRuntimeConfig.HTTPHostname, "/")
	}
	pageURL := base + r.URL.Path

	var buf bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(r.Context()))
	if err != nil {
		log.Printf("parse template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	content := struct{ To, From, Subject, URL string }{
		To:      user.Email.String,
		From:    email.SourceEmail,
		Subject: "Website Update Notification",
		URL:     pageURL,
	}
	if err := tmpl.Execute(&buf, content); err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := provider.Send(r.Context(), user.Email.String, content.Subject, buf.String()); err != nil {
		log.Printf("send email: %v", err)
	}
	http.Redirect(w, r, "/admin/email/template", http.StatusSeeOther)
}
