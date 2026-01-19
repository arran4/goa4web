package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

const AdminPasswordResetListPageTmpl handlers.Page = "admin/passwordResetList.gohtml"

type AdminPasswordResetListPageData struct {
	Rows       []*db.AdminListPasswordResetsRow
	Status     string
	Page       int
	TotalPages int
}

func adminPasswordResetListPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Password Resets"

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id: %w", err))
			return
		}

		switch action {
		case "approve":
			if err := cd.AdminApprovePasswordReset(int32(id)); err != nil {
				handlers.RenderErrorPage(w, r, fmt.Errorf("approve failed: %w", err))
				return
			}
		case "deny":
			if err := cd.AdminDenyPasswordReset(int32(id)); err != nil {
				handlers.RenderErrorPage(w, r, fmt.Errorf("deny failed: %w", err))
				return
			}
		}
		http.Redirect(w, r, r.URL.Path+"?"+r.URL.RawQuery, http.StatusSeeOther)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		status = "pending"
	}
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	pageStr := r.FormValue("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	const pageSize = 20
	offset := int32((page - 1) * pageSize)

	rows, count, err := cd.AdminListPasswordResets(statusPtr, pageSize, offset)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list failed: %w", err))
		return
	}

	totalPages := int((count + pageSize - 1) / pageSize)

	AdminPasswordResetListPageTmpl.Handle(w, r, &AdminPasswordResetListPageData{
		Rows:       rows,
		Status:     status,
		Page:       page,
		TotalPages: totalPages,
	})
}
