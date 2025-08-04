package user

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminPendingUsersPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	rows, err := queries.AdminListPendingUsers(r.Context())
	if err != nil && err != sql.ErrNoRows {
		log.Printf("list pending users: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := struct {
		*common.CoreData
		Rows []*db.AdminListPendingUsersRow
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Rows:     rows,
	}
	handlers.TemplateHandler(w, r, "pendingUsersPage.gohtml", data)
}

func adminPendingUsersApprove(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid := r.PostFormValue("uid")
	var id int32
	fmt.Sscanf(uid, "%d", &id)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/pending",
	}
	if id == 0 {
		data.Errors = append(data.Errors, "invalid id")
	} else {
		if err := queries.SystemCreateUserRole(r.Context(), db.SystemCreateUserRoleParams{UsersIdusers: id, Name: "user"}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("add role: %w", err).Error())
		} else {
			data.Messages = append(data.Messages, "User approved")
		}

	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func adminPendingUsersReject(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid := r.PostFormValue("uid")
	reason := r.PostFormValue("reason")
	var id int32
	fmt.Sscanf(uid, "%d", &id)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/pending",
	}
	if id == 0 {
		data.Errors = append(data.Errors, "invalid id")
	} else {
		if err := queries.SystemCreateUserRole(r.Context(), db.SystemCreateUserRoleParams{UsersIdusers: id, Name: "rejected"}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("add role:%w", err).Error())
		} else {
			data.Messages = append(data.Messages, "user rejected")
		}
		if reason != "" {
			if err := queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: id, Comment: reason}); err != nil {
				log.Printf("insert admin user comment: %v", err)
			} else {
				data.Messages = append(data.Messages, "comment recorded")
			}
		}

	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
