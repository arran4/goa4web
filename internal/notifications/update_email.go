package notifications

import (
	"context"

	"github.com/arran4/goa4web/config"
	coretemplates "github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	ttemplate "text/template"
)

// GetUpdateEmailText returns the update email text template after applying any
// database overrides.
func GetUpdateEmailText(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig) string {
	tmpls := coretemplates.GetCompiledEmailTextTemplates(map[string]any{}, coretemplates.WithDir(cfg.TemplatesDir))
	b, err := renderTemplate[*ttemplate.Template](ctx, q, EmailTextTemplateFilenameGenerator("updateEmail"), nil, tmpls, TextTemplatesNew)
	if err != nil {
		return ""
	}
	return string(b)
}
