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
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AcceptRequestTask accepts a queued request.
type AcceptRequestTask struct{ tasks.TaskString }

var acceptRequestTask = &AcceptRequestTask{TaskString: TaskAccept}

// RejectRequestTask rejects a queued request.
type RejectRequestTask struct{ tasks.TaskString }

var rejectRequestTask = &RejectRequestTask{TaskString: TaskReject}

// QueryRequestTask asks for more information about a request.
type QueryRequestTask struct{ tasks.TaskString }

var queryRequestTask = &QueryRequestTask{TaskString: TaskQuery}

var _ tasks.Task = (*AcceptRequestTask)(nil)
var _ tasks.Task = (*RejectRequestTask)(nil)
var _ tasks.Task = (*QueryRequestTask)(nil)

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
	handlers.TemplateHandler(w, r, "admin/requestQueuePage.gohtml", data)
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
	handlers.TemplateHandler(w, r, "admin/requestArchivePage.gohtml", data)
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
	handlers.TemplateHandler(w, r, "admin/requestPage.gohtml", data)
}

func adminRequestAddCommentPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	comment := r.PostFormValue("comment")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := struct {
		*common.CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     fmt.Sprintf("/admin/request/%d", id),
	}
	if comment == "" || id == 0 {
		data.Errors = append(data.Errors, "invalid")
	} else {
		_ = queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: comment})
	}
	handlers.TemplateHandler(w, r, "admin/runTaskPage.gohtml", data)
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
	_ = queries.UpdateAdminRequestStatus(r.Context(), db.UpdateAdminRequestStatusParams{Status: status, ID: int32(id)})
	auto := fmt.Sprintf("status changed to %s", status)
	_ = queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: auto})
	if comment != "" {
		_ = queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: comment})
	}
	_ = queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: req.UsersIdusers, Comment: auto})
	data := struct {
		*common.CoreData
		Back string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/requests",
	}
	handlers.TemplateHandler(w, r, "admin/runTaskPage.gohtml", data)
}

func (AcceptRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "accepted")
	return nil
}
func (RejectRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "rejected")
	return nil
}
func (QueryRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "query")
	return nil
}
