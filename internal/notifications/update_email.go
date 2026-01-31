package notifications

import (
	"context"
	"strings"

	ttemplate "text/template"

	"github.com/arran4/goa4web/config"
	coretemplates "github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

// GetUpdateEmailText returns the update email text template after applying any
// database overrides.
func GetUpdateEmailText(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig) string {
	tmpls := coretemplates.GetCompiledEmailTextTemplates(map[string]any{
		"truncateWords": func(i int, s string) string {
			words := strings.Fields(s)
			if len(words) > i {
				return strings.Join(words[:i], " ") + "..."
			}
			return s
		},
	}, coretemplates.WithDir(cfg.TemplatesDir))
	b, err := renderTemplate[*ttemplate.Template](ctx, q, EmailTextTemplateFilenameGenerator("updateEmail"), nil, tmpls, TextTemplatesNew)
	if err != nil {
		return ""
	}
	return string(b)
}
