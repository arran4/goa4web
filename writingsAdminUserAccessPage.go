package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func writingsAdminUserAccessPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		ApprovedUsers []*GetAllWritingApprovalsRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	approvedUserRows, err := queries.GetAllWritingApprovals(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingApprovals Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.ApprovedUsers = approvedUserRows

	CustomWritingsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminUserAccessPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAdminUserAccessAllowActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	username := r.PostFormValue("username")
	where := r.PostFormValue("where")
	level := r.PostFormValue("level")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section: sql.NullString{
			String: where,
			Valid:  true,
		},
		Level: sql.NullString{
			String: level,
			Valid:  true,
		},
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func writingsAdminUserAccessAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	wid, _ := strconv.Atoi(r.PostFormValue("wid"))
	username := r.PostFormValue("username")
	readdoc, _ := strconv.ParseBool(r.PostFormValue("readdoc"))
	editdoc, _ := strconv.ParseBool(r.PostFormValue("editdoc"))
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.CreateWritingApproval(r.Context(), CreateWritingApprovalParams{
		WritingIdwriting: int32(wid),
		UsersIdusers:     int32(u.Idusers),
		Readdoc:          sql.NullBool{Valid: true, Bool: readdoc},
		Editdoc:          sql.NullBool{Valid: true, Bool: editdoc},
	}); err != nil {
		log.Printf("createWritingApproval Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
func writingsAdminUserAccessUpdateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	uid, _ := strconv.Atoi(r.PostFormValue("uid"))
	wid, _ := strconv.Atoi(r.PostFormValue("wid"))
	readdoc, _ := strconv.ParseBool(r.PostFormValue("readdoc"))
	editdoc, _ := strconv.ParseBool(r.PostFormValue("editdoc"))

	if err := queries.UpdateWritingApproval(r.Context(), UpdateWritingApprovalParams{
		WritingIdwriting: int32(wid),
		UsersIdusers:     int32(uid),
		Readdoc:          sql.NullBool{Valid: true, Bool: readdoc},
		Editdoc:          sql.NullBool{Valid: true, Bool: editdoc},
	}); err != nil {
		log.Printf("createWritingApproval Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func writingsAdminUserAccessRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	uid, _ := strconv.Atoi(r.PostFormValue("uid"))
	wid, _ := strconv.Atoi(r.PostFormValue("wid"))

	if err := queries.DeleteWritingApproval(r.Context(), DeleteWritingApprovalParams{
		WritingIdwriting: int32(wid),
		UsersIdusers:     int32(uid),
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
