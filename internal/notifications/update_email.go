package notifications

import (
	"context"

	coretemplates "github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	ttemplate "text/template"
)

// GetUpdateEmailText returns the update email text template after applying any
// database overrides.
func GetUpdateEmailText(ctx context.Context, q db.Querier) string {
	tmpls := coretemplates.GetCompiledEmailTextTemplates(map[string]any{})
	b, err := renderTemplate[*ttemplate.Template](ctx, q, EmailTextTemplateFilenameGenerator("updateEmail"), nil, tmpls, TextTemplatesNew)
	if err != nil {
		return ""
	}
	return string(b)
}
