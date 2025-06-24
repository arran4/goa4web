package goa4web

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func linkerAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type PermissionUser struct {
		*Permission
		Username sql.NullString
		Email    sql.NullString
	}

	type Data struct {
		*CoreData
		UserLevels []*PermissionUser
		Search     string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.GetUsersPermissions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getUsersPermissions Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	var perms []*PermissionUser
	for _, p := range rows {
		u, err := queries.GetUserById(r.Context(), p.UsersIdusers)
		if err != nil {
			log.Printf("GetUserById Error: %s", err)
			continue
		}
		perms = append(perms, &PermissionUser{Permission: p, Username: u.Username, Email: u.Email})
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*PermissionUser
		for _, row := range perms {
			if strings.Contains(strings.ToLower(row.Username.String), q) {
				filtered = append(filtered, row)
			}
		}
		perms = filtered
	}
	data.UserLevels = perms

	CustomLinkerIndex(data.CoreData, r)
	if err := renderTemplate(w, r, "linkerAdminUserLevelsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	usernames := r.PostFormValue("usernames")
	where := r.PostFormValue("where")
	level := r.PostFormValue("level")
	fields := strings.FieldsFunc(usernames, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	for _, username := range fields {
		if username == "" {
			continue
		}
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("GetUserByUsername Error: %s", err)
			continue
		}
		if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
			UsersIdusers: u.Idusers,
			Section:      sql.NullString{String: where, Valid: true},
			Level:        sql.NullString{String: level, Valid: true},
		}); err != nil {
			log.Printf("permissionUserAllow Error: %s", err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	r.ParseForm()
	ids := r.Form["permids"]
	if len(ids) == 0 {
		if id := r.PostFormValue("permid"); id != "" {
			ids = append(ids, id)
		}
	}
	for _, idStr := range ids {
		permid, _ := strconv.Atoi(idStr)
		if err := queries.PermissionUserDisallow(r.Context(), int32(permid)); err != nil {
			log.Printf("permissionUserDisallow Error: %s", err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}
