package goa4web

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// adminEmailTemplatePage allows administrators to edit the update email template.
func adminEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
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
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Body:     b,
		Preview:  preview,
		Error:    r.URL.Query().Get("error"),
	}

	if err := renderTemplate(w, r, "emailTemplatePage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminEmailTemplateSaveActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body := r.PostFormValue("body")
	q := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := q.SetTemplateOverride(r.Context(), "updateEmail", body); err != nil {
		log.Printf("db save template: %v", err)
	}
	http.Redirect(w, r, "/admin/email/template", http.StatusSeeOther)
}

func adminEmailTemplateTestActionPage(w http.ResponseWriter, r *http.Request) {
	provider := getEmailProvider()
	if provider == nil {
		q := url.QueryEscape(errMailNotConfigured)
		http.Redirect(w, r, "/admin/email/template?error="+q, http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	user, err := queries.GetUserById(r.Context(), cd.UserID)
	if err != nil {
		log.Printf("get user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
