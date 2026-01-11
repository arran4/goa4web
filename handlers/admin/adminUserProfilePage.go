package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserProfilePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("User %s", user.Username.String)
	AdminUserProfilePageTmpl.Handle(w, r, struct{}{})
}

const AdminUserProfilePageTmpl handlers.Page = "admin/adminUserPage.gohtml"

func adminUserAddCommentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	back := "/admin/user"
	if user != nil {
		back = fmt.Sprintf("/admin/user/%d", user.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	comment := r.PostFormValue("comment")
	if user == nil {
		data.Errors = append(data.Errors, "invalid user id")
	} else if strings.TrimSpace(comment) == "" {
		data.Errors = append(data.Errors, "empty comment")
	} else {
		if err := cd.Queries().InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: user.Idusers, Comment: comment}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		} else {
			data.Messages = append(data.Messages, "comment added")
		}
	}
	RunTaskPageTmpl.Handle(w, r, data)
}
