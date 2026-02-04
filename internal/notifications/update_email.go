package notifications

import (
	"context"

	ttemplate "text/template"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	coretemplates "github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

// GetUpdateEmailText returns the update email text template after applying any
// database overrides.
func GetUpdateEmailText(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig) string {
	tmpls := coretemplates.GetCompiledEmailTextTemplates(common.GetTemplateFuncs(), coretemplates.WithDir(cfg.TemplatesDir), coretemplates.WithSilence(cfg.Silent))
	b, err := renderTemplate[*ttemplate.Template](ctx, q, EmailTextTemplateFilenameGenerator("updateEmail"), nil, tmpls, TextTemplatesNew)
	if err != nil {
		return ""
	}
	return string(b)
}
