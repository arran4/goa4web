package goa4web

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/mux"
)

func writingsCategoryPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		CategoryID int32
		UserLevels []*PermissionWithUser
	}
	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	vars := mux.Vars(r)
	cid, _ := strconv.Atoi(vars["category"])
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.GetPermissionsBySectionWithUsers(r.Context(), fmt.Sprintf("writing:%d", cid))
	if err != nil && err != sql.ErrNoRows {
		log.Printf("getPermissions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := Data{CoreData: cd, CategoryID: int32(cid), UserLevels: rows}
	CustomWritingsIndex(cd, r)
	if err := templates.RenderTemplate(w, "categoryPermissionsPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsCategoryPermissionsAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	cid, _ := strconv.Atoi(vars["category"])
	username := r.PostFormValue("username")
	level := r.PostFormValue("level")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section:      sql.NullString{Valid: true, String: fmt.Sprintf("writing:%d", cid)},
		Level:        sql.NullString{Valid: true, String: level},
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func writingsCategoryPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	permid, _ := strconv.Atoi(r.PostFormValue("permid"))
	if err := queries.PermissionUserDisallow(r.Context(), int32(permid)); err != nil {
		log.Printf("permissionUserDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}
