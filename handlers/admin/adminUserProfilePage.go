package admin

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserProfilePage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	user, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	emails, _ := queries.GetUserEmailsByUserID(r.Context(), int32(id))
	comments, _ := queries.ListAdminUserComments(r.Context(), int32(id))
	roles, _ := queries.GetPermissionsByUserID(r.Context(), int32(id))
	stats, _ := queries.UserPostCountsByID(r.Context(), int32(id))
	bm, _ := queries.GetBookmarksForUser(r.Context(), int32(id))
	var bmSize int
	if bm != nil {
		list := strings.TrimSpace(bm.List.String)
		if list != "" {
			bmSize = len(strings.Split(list, "\n"))
		}
	}
	grants, _ := queries.ListGrantsByUserID(r.Context(), sql.NullInt32{Int32: int32(id), Valid: true})
	cd.PageTitle = fmt.Sprintf("User %s", user.Username.String)
	data := struct {
		*common.CoreData
		User         *db.User
		Emails       []*db.UserEmail
		Comments     []*db.AdminUserComment
		Roles        []*db.GetPermissionsByUserIDRow
		Stats        *db.UserPostCountsByIDRow
		BookmarkSize int
		Grants       []*db.Grant
	}{
		CoreData:     cd,
		User:         &db.User{Idusers: user.Idusers, Username: user.Username},
		Emails:       emails,
		Comments:     comments,
		Roles:        roles,
		Stats:        stats,
		BookmarkSize: bmSize,
		Grants:       grants,
	}
	handlers.TemplateHandler(w, r, "userProfile.gohtml", data)
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
