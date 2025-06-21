package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
)

func adminUsersPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type SectionGroup struct {
		Section string
		Rows    []*PermissionWithUser
	}
	type Data struct {
		*CoreData
		Sections []SectionGroup
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.GetPermissionsWithUsers(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Section.String == rows[j].Section.String {
			return rows[i].Username.String < rows[j].Username.String
		}
		return rows[i].Section.String < rows[j].Section.String
	})
	var groups []SectionGroup
	for _, r := range rows {
		sec := r.Section.String
		if len(groups) == 0 || groups[len(groups)-1].Section != sec {
			groups = append(groups, SectionGroup{Section: sec})
		}
		groups[len(groups)-1].Rows = append(groups[len(groups)-1].Rows, r)
	}
	data.Sections = groups

	err = renderTemplate(w, r, "adminUsersPermissionsPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	username := r.PostFormValue("username")
	where := r.PostFormValue("where")
	level := r.PostFormValue("level")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users/permissions",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section: sql.NullString{
			String: where,
			Valid:  true,
		},
		Level: sql.NullString{
			String: level,
			Valid:  true,
		},
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}
	err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users/permissions",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.PermissionUserDisallow(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUsersPermissionsUpdatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	permid := r.PostFormValue("permid")
	level := r.PostFormValue("level")
	where := r.PostFormValue("where")

	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users/permissions",
	}

	if id, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.UpdatePermission(r.Context(), UpdatePermissionParams{
		ID:      int32(id),
		Section: sql.NullString{String: where, Valid: true},
		Level:   sql.NullString{String: level, Valid: true},
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("UpdatePermission: %w", err).Error())
	}

	if err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
