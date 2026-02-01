package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AdminExternalLinkDetailsPage displays details of a single external link.
func AdminExternalLinkDetailsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Link          *db.ExternalLink
		ResultSummary string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasAdminRole() {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	queries := cd.Queries()
	link, err := queries.GetExternalLinkByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}
		handlers.RenderErrorPage(w, r, fmt.Errorf("fetching external link: %w", err))
		return
	}

	cd.PageTitle = fmt.Sprintf("External Link %d", link.ID)
	cd.SetCurrentPage(&AdminExternalLinkDetailsPageBreadcrumb{LinkID: link.ID})

	data := Data{
		Link: link,
	}

	// Result Summary Logic (similar to list page)
	action := r.URL.Query().Get(externalLinksActionQueryParam)
	successCount := queryIntValue(r, externalLinksSuccessQueryParam)
	failureCount := queryIntValue(r, externalLinksFailureQueryParam)
	if action != "" {
		actionLabel := action
		switch action {
		case externalLinksActionRefresh:
			actionLabel = "Refreshed"
		case externalLinksActionDelete:
			actionLabel = "Deleted"
		}
		data.ResultSummary = fmt.Sprintf("%s external link: %d succeeded, %d failed.", actionLabel, successCount, failureCount)
	}

	AdminExternalLinkDetailsPageTmpl.Handle(w, r, data)
}

const AdminExternalLinkDetailsPageTmpl tasks.Template = "admin/externalLinkDetailsPage.gohtml"
