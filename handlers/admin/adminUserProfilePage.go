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
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("User %s", user.Username.String)
	handlers.TemplateHandler(w, r, "userProfile.gohtml", cd)
}

func adminUserAddCommentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	user := cd.CurrentProfileUser()
	back := "/admin/user"
	if user != nil {
		back = fmt.Sprintf("/admin/user/%d", user.Idusers)
	}
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     back,
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
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
