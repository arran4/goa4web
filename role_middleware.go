package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
)

// roleAllowed checks if the current request has one of the provided roles.
func roleAllowed(r *http.Request, roles ...string) bool {
	cd, ok := r.Context().Value(ContextValues("coreData")).(*CoreData)
	if ok && cd != nil {
		for _, lvl := range roles {
			if cd.HasRole(lvl) {
				return true
			}
		}
		return false
	}

	user, uok := r.Context().Value(ContextValues("user")).(*User)
	queries, qok := r.Context().Value(ContextValues("queries")).(*Queries)
	if !uok || !qok {
		return false
	}
	section := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
	perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(r.Context(), GetPermissionsByUserIdAndSectionAndSectionAllParams{
		UsersIdusers: user.Idusers,
		Section:      sql.NullString{String: section, Valid: true},
	})
	if err != nil || !perm.Level.Valid {
		return false
	}
	cd = &CoreData{SecurityLevel: perm.Level.String}
	for _, lvl := range roles {
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !roleAllowed(r, roles...) {
				err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "adminNoAccessPage.gohtml", r.Context().Value(ContextValues("coreData")).(*CoreData))
				if err != nil {
					log.Printf("Template Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
