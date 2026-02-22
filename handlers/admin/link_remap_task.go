package admin

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// ApplyLinkRemapTask updates site news URLs based on a mapping.
type ApplyLinkRemapTask struct{ tasks.TaskString }

var applyLinkRemapTask = &ApplyLinkRemapTask{TaskString: TaskUpdate}

var _ tasks.Task = (*ApplyLinkRemapTask)(nil)
var _ tasks.AuditableTask = (*ApplyLinkRemapTask)(nil)

func (ApplyLinkRemapTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	csvText := r.PostFormValue("csv")
	reader := csv.NewReader(strings.NewReader(csvText))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read csv fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	q := cd.Queries()
	ctx := r.Context()
	for i, rec := range records {
		if i == 0 && len(rec) > 0 && strings.EqualFold(rec[0], "internal reference") {
			continue
		}
		if len(rec) < 3 || rec[2] == "" {
			continue
		}
		parts := strings.SplitN(rec[0], ":", 2)
		if len(parts) != 2 {
			continue
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		switch parts[0] {
		case "site_news":
			if err := q.AdminReplaceSiteNewsURL(ctx, db.AdminReplaceSiteNewsURLParams{OldUrl: rec[1], NewUrl: rec[2], ID: int32(id)}); err != nil {
				return fmt.Errorf("update news %d fail %w", id, handlers.ErrRedirectOnSamePageHandler(err))
			}
			if err := q.AdminDeleteExternalLinkByURL(ctx, rec[1]); err != nil {
				return fmt.Errorf("cleanup external link %q fail %w", rec[1], handlers.ErrRedirectOnSamePageHandler(err))
			}
		default:
			continue
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/link-discovery"}
}

func (ApplyLinkRemapTask) AuditRecord(data map[string]any) string {
	return "applied link remap"
}
