package user

import (
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
)

func adminSessionsPage(w http.ResponseWriter, r *http.Request) {
	AdminSessionsPage.Handle(w, r, struct{}{})
}

const AdminSessionsPage handlers.Page = "admin/sessionsPage.gohtml"

func adminSessionsDeletePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	sid := r.PostFormValue("sid")
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/sessions",
	}
	if sid == "" {
		data.Errors = append(data.Errors, "missing sid")
	} else {
		sm := cd.SessionManager()
		if err := sm.DeleteSessionByID(r.Context(), sid); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}
	AdminRunTaskPage.Handle(w, r, data)
}
