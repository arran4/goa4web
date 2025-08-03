package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

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
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/user/" + idStr,
	}
	comment := r.PostFormValue("comment")
	if id == 0 {
		data.Errors = append(data.Errors, "invalid user id")
	} else if strings.TrimSpace(comment) == "" {
		data.Errors = append(data.Errors, "empty comment")
	} else {
		queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
		if err := queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: int32(id), Comment: comment}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		} else {
			data.Messages = append(data.Messages, "comment added")
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
