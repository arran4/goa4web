package admin

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminRequestQueuePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Requests"
	AdminRequestQueuePageTmpl.Handle(w, r, struct{}{})
}

const AdminRequestQueuePageTmpl tasks.Template = "admin/requestQueuePage.gohtml"

func AdminRequestArchivePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Request Archive"
	AdminRequestArchivePageTmpl.Handle(w, r, struct{}{})
}

const AdminRequestArchivePageTmpl tasks.Template = "admin/requestArchivePage.gohtml"

func adminRequestPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	id := cd.CurrentRequestID()
	if id == 0 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("not found"))
		return
	}
	if req := cd.CurrentRequest(); req != nil {
		cd.SetCurrentProfileUserID(req.UsersIdusers)
	}
	cd.PageTitle = fmt.Sprintf("Request %d", id)
	AdminRequestPageTmpl.Handle(w, r, struct{}{})
}

const AdminRequestPageTmpl tasks.Template = "admin/requestPage.gohtml"

func adminRequestAddCommentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	req := cd.CurrentRequest()
	var id int32
	if req != nil {
		id = req.ID
		cd.SetCurrentProfileUserID(req.UsersIdusers)
	}
	comment := r.PostFormValue("comment")
	cd.PageTitle = "Add Comment"
	queries := cd.Queries()
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: fmt.Sprintf("/admin/request/%d", id),
	}
	if comment == "" {
		data.Errors = append(data.Errors, "invalid")
	} else if err := queries.AdminInsertRequestComment(r.Context(), db.AdminInsertRequestCommentParams{RequestID: id, Comment: comment}); err != nil {
		data.Errors = append(data.Errors, err.Error())
	} else {
		data.Messages = append(data.Messages, "comment added")
	}
	AdminRequestPageTmpl.Handle(w, r, data)
}

func handleRequestAction(w http.ResponseWriter, r *http.Request, status string) {
	comment := r.PostFormValue("comment")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	if cd == nil || !cd.HasAdminRole() {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	req := cd.CurrentRequest()
	if req == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Request %d", req.ID)
	queries := cd.Queries()
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/requests",
	}

	var auto string

	if err := queries.AdminUpdateRequestStatus(r.Context(), db.AdminUpdateRequestStatusParams{Status: status, ID: req.ID}); err != nil {
		data.Errors = append(data.Errors, err.Error())
	} else {
		auto = fmt.Sprintf("status changed to %s", status)
		data.Messages = append(data.Messages, auto)
		if err := queries.AdminInsertRequestComment(r.Context(), db.AdminInsertRequestCommentParams{RequestID: req.ID, Comment: auto}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
		if comment != "" {
			if err := queries.AdminInsertRequestComment(r.Context(), db.AdminInsertRequestCommentParams{RequestID: req.ID, Comment: comment}); err != nil {
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
			evt.Data["RequestID"] = int(req.ID)
			evt.Data["Status"] = status
		}
	}
	data.Messages = append(data.Messages, auto)
	RunTaskPageTmpl.Handle(w, r, data)
}

func requestAuditSummary(action string, data map[string]any) string {
	id, _ := data["RequestID"].(int)
	if id != 0 {
		return fmt.Sprintf("request %d %s", id, action)
	}
	return "request " + action
}
