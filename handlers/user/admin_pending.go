package user

import (
	"database/sql"
	"fmt"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	emailutil "github.com/arran4/goa4web/internal/notifications"
)

func adminPendingUsersPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	rows, err := queries.ListPendingUsers(r.Context())
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*handlers.CoreData
		Rows []*db.ListPendingUsersRow
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Rows:     rows,
	}
	handlers.TemplateHandler(w, r, "admin/pendingUsersPage.gohtml", data)
}

func adminPendingUsersApprove(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	uid := r.PostFormValue("uid")
	var id int32
	fmt.Sscanf(uid, "%d", &id)
	data := struct {
		*handlers.CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Back:     "/admin/users/pending",
	}
	if id == 0 {
		data.Errors = append(data.Errors, "invalid id")
	} else {
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{UsersIdusers: id, Name: "user"}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("add role: %w", err).Error())
		}
		if u, err := queries.GetUserById(r.Context(), id); err == nil && u.Email.Valid {
			_ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, id, u.Email.String, "", "user approved", nil)
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func adminPendingUsersReject(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	uid := r.PostFormValue("uid")
	reason := r.PostFormValue("reason")
	var id int32
	fmt.Sscanf(uid, "%d", &id)
	data := struct {
		*handlers.CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Back:     "/admin/users/pending",
	}
	if id == 0 {
		data.Errors = append(data.Errors, "invalid id")
	} else {
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{UsersIdusers: id, Name: "rejected"}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("add role:%w", err).Error())
		}
		if reason != "" {
			_ = queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: id, Comment: reason})
		}
		if u, err := queries.GetUserById(r.Context(), id); err == nil && u.Email.Valid {
			item := struct{ Reason string }{Reason: reason}
			_ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, id, u.Email.String, "", "user rejected", item)
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
