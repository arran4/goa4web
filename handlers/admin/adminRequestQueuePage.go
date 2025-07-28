package admin

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func AdminRequestQueuePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	rows, err := queries.ListPendingAdminRequests(r.Context())
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	type Row struct {
		*db.AdminRequestQueue
		Username string
	}
	data := struct {
		*common.CoreData
		Rows []Row
	}{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	for _, row := range rows {
		user, err := queries.GetUserById(r.Context(), row.UsersIdusers)
		if err != nil {
			continue
		}
		data.Rows = append(data.Rows, Row{row, user.Username.String})
	}
	handlers.TemplateHandler(w, r, "requestQueuePage.gohtml", data)
}

func AdminRequestArchivePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	rows, err := queries.ListArchivedAdminRequests(r.Context())
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	type Row struct {
		*db.AdminRequestQueue
		Username string
	}
	data := struct {
		*common.CoreData
		Rows []Row
	}{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	for _, row := range rows {
		user, err := queries.GetUserById(r.Context(), row.UsersIdusers)
		if err != nil {
			continue
		}
		data.Rows = append(data.Rows, Row{row, user.Username.String})
	}
	handlers.TemplateHandler(w, r, "requestArchivePage.gohtml", data)
}

func adminRequestPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	req, err := queries.GetAdminRequestByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	comments, _ := queries.ListAdminRequestComments(r.Context(), int32(id))
	user, _ := queries.GetUserById(r.Context(), req.UsersIdusers)
	data := struct {
		*common.CoreData
		Req      *db.AdminRequestQueue
		User     *db.GetUserByIdRow
		Comments []*db.AdminRequestComment
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Req:      req,
		User:     user,
		Comments: comments,
	}
	handlers.TemplateHandler(w, r, "requestPage.gohtml", data)
}

func adminRequestAddCommentPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	comment := r.PostFormValue("comment")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     fmt.Sprintf("/admin/request/%d", id),
	}
	if comment == "" || id == 0 {
		data.Errors = append(data.Errors, "invalid")
	} else if err := queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: comment}); err != nil {
		data.Errors = append(data.Errors, err.Error())
	} else {
		data.Messages = append(data.Messages, "comment added")
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func handleRequestAction(w http.ResponseWriter, r *http.Request, status string) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	comment := r.PostFormValue("comment")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	req, err := queries.GetAdminRequestByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/requests",
	}

	if err := queries.UpdateAdminRequestStatus(r.Context(), db.UpdateAdminRequestStatusParams{Status: status, ID: int32(id)}); err != nil {
		data.Errors = append(data.Errors, err.Error())
	} else {
		auto := fmt.Sprintf("status changed to %s", status)
		data.Messages = append(data.Messages, auto)
		if err := queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: auto}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
		if comment != "" {
			if err := queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: comment}); err != nil {
				data.Errors = append(data.Errors, err.Error())
			}
		}
		if err := queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: req.UsersIdusers, Comment: auto}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["RequestID"] = id
			evt.Data["Status"] = status
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func requestAuditSummary(action string, data map[string]any) string {
	id, _ := data["RequestID"].(int)
	if id != 0 {
		return fmt.Sprintf("request %d %s", id, action)
	}
	return "request " + action
}
