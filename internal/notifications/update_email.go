package notifications

import (
	"context"

	coretemplates "github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"
	htemplate "html/template"
)

// GetUpdateEmailText returns the update email text template after applying any
// database overrides.
func GetUpdateEmailText(ctx context.Context, q *db.Queries) string {
	tmpls := coretemplates.GetCompiledEmailTextTemplates(map[string]any{})
	b, err := renderTemplate[*htemplate.Template](ctx, q, EmailTextTemplateFilenameGenerator("updateEmail"), nil, tmpls, HTMLTemplatesNew)
	if err != nil {
		return ""
	}
	return string(b)
}
