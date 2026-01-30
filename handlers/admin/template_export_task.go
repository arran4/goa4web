package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	coretemplates "github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// ExportTemplatesTask streams embedded templates as an archive.
type ExportTemplatesTask struct{ tasks.TaskString }

var exportTemplatesTask = &ExportTemplatesTask{TaskString: TaskExportTemplates}

var _ tasks.Task = (*ExportTemplatesTask)(nil)
var _ tasks.AuditableTask = (*ExportTemplatesTask)(nil)

// Action streams the selected templates as a zip or tar archive.
func (ExportTemplatesTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	setValue := strings.ToLower(strings.TrimSpace(r.PostFormValue("set")))
	format := strings.ToLower(strings.TrimSpace(r.PostFormValue("format")))
	if setValue == "" {
		setValue = string(coretemplates.TemplateSetSite)
	}
	if format == "" {
		format = "zip"
	}
	set, err := parseTemplateSet(setValue)
	if err != nil {
		return fmt.Errorf("invalid template set: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if format != "zip" && format != "tar" {
		return fmt.Errorf("invalid archive format: %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("unknown format %q", format)))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["TemplateSet"] = string(set)
			evt.Data["Format"] = format
		}
	}

	filename := fmt.Sprintf("%s-templates-%s.%s", set, time.Now().UTC().Format("20060102"), format)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch format {
		case "zip":
			w.Header().Set("Content-Type", "application/zip")
		case "tar":
			w.Header().Set("Content-Type", "application/x-tar")
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		if err := coretemplates.ArchiveTemplates(w, format, []coretemplates.TemplateSet{set}); err != nil {
			handlers.RenderErrorPage(w, r, fmt.Errorf("export templates: %w", err))
		}
	})
}

// AuditRecord summarises exporting embedded templates.
func (ExportTemplatesTask) AuditRecord(data map[string]any) string {
	set, _ := data["TemplateSet"].(string)
	format, _ := data["Format"].(string)
	switch {
	case set != "" && format != "":
		return fmt.Sprintf("exported %s templates as %s", set, format)
	case set != "":
		return fmt.Sprintf("exported %s templates", set)
	default:
		return "exported templates"
	}
}

func parseTemplateSet(value string) (coretemplates.TemplateSet, error) {
	switch coretemplates.TemplateSet(value) {
	case coretemplates.TemplateSetSite, coretemplates.TemplateSetEmail:
		return coretemplates.TemplateSet(value), nil
	default:
		return "", fmt.Errorf("unsupported set %q", value)
	}
}
