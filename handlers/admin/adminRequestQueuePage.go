package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminRequestQueuePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Requests"
	handlers.TemplateHandler(w, r, "requestQueuePage.gohtml", cd)
}

func AdminRequestArchivePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Request Archive"
	handlers.TemplateHandler(w, r, "requestArchivePage.gohtml", cd)
}

func adminRequestPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	id := cd.CurrentRequestID()
	if id == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("Request %d", id)
	handlers.TemplateHandler(w, r, "requestPage.gohtml", cd)
}

func adminRequestAddCommentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	req := cd.CurrentRequest()
	var id int32
	if req != nil {
		id = req.ID
	}
	comment := r.PostFormValue("comment")
	cd.PageTitle = "Add Comment"
	queries := cd.Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     fmt.Sprintf("/admin/request/%d", id),
	}
	if comment == "" {
		data.Errors = append(data.Errors, "invalid")
	} else if err := queries.AdminInsertRequestComment(r.Context(), db.AdminInsertRequestCommentParams{RequestID: id, Comment: comment}); err != nil {
		data.Errors = append(data.Errors, err.Error())
	} else {
		data.Messages = append(data.Messages, "comment added")
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func handleRequestAction(w http.ResponseWriter, r *http.Request, status string) {
	comment := r.PostFormValue("comment")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	req := cd.CurrentRequest()
	if req == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("Request %d", req.ID)
	queries := cd.Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin/requests",
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
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func requestAuditSummary(action string, data map[string]any) string {
	id, _ := data["RequestID"].(int)
	if id != 0 {
		return fmt.Sprintf("request %d %s", id, action)
	}
	return "request " + action
}
