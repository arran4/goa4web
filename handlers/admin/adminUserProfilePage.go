package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func adminUserProfilePage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	user, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	emails, _ := queries.GetUserEmailsByUserID(r.Context(), int32(id))
	comments, _ := queries.ListAdminUserComments(r.Context(), int32(id))
	data := struct {
		*common.CoreData
		User     *db.User
		Emails   []*db.UserEmail
		Comments []*db.AdminUserComment
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Emails:   emails,
		Comments: comments,
	}
	handlers.TemplateHandler(w, r, "admin/userProfile.gohtml", data)
}

func adminUserAddCommentPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	data := struct {
		*common.CoreData
		Errors []string
		Back   string
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
		queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
		if err := queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: int32(id), Comment: comment}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}
	handlers.TemplateHandler(w, r, "admin/runTaskPage.gohtml", data)
}
