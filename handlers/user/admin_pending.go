package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminPendingUsersPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	rows, err := queries.AdminListPendingUsers(r.Context())
	if err != nil && err != sql.ErrNoRows {
		log.Printf("list pending users: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	emailRows, err := queries.GetVerifiedUserEmails(r.Context())
	if err != nil {
		log.Printf("list pending user emails: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	emailsByUser := db.EmailsByUserID(emailRows)
	type pendingUser struct {
		*db.AdminListPendingUsersRow
		Emails       []string
		PrimaryEmail string
	}
	pending := make([]pendingUser, 0, len(rows))
	for _, row := range rows {
		emails := emailsByUser[row.Idusers]
		pending = append(pending, pendingUser{
			AdminListPendingUsersRow: row,
			Emails:                   emails,
			PrimaryEmail:             db.PrimaryEmail(emails),
		})
	}
	data := struct {
		Rows []pendingUser
	}{
		Rows: pending,
	}
	AdminPendingUsersPage.Handle(w, r, data)
}

const AdminPendingUsersPage handlers.Page = "admin/pendingUsersPage.gohtml"

func adminPendingUsersApprove(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	uid := r.PostFormValue("uid")
	var id int32
	fmt.Sscanf(uid, "%d", &id)
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/users/pending",
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
	AdminRunTaskPage.Handle(w, r, data)
}

func adminPendingUsersReject(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	uid := r.PostFormValue("uid")
	reason := r.PostFormValue("reason")
	var id int32
	fmt.Sscanf(uid, "%d", &id)
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/users/pending",
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
	AdminRunTaskPage.Handle(w, r, data)
}
