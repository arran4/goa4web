package admin

import (
	"bytes"
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"text/template"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
)

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
