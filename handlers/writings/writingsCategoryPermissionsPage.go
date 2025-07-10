package writings

import (
	"database/sql"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func CategoryPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		CategoryID int32
		UserLevels []*db.PermissionWithUser
	}
	cd := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData)
	vars := mux.Vars(r)
	cid, _ := strconv.Atoi(vars["category"])
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.GetPermissionsBySectionWithUsers(r.Context(), fmt.Sprintf("writing:%d", cid))
	if err != nil && err != sql.ErrNoRows {
		log.Printf("getPermissions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := Data{CoreData: cd, CategoryID: int32(cid), UserLevels: rows}
	CustomWritingsIndex(cd, r)
	if err := templates.RenderTemplate(w, "categoryPermissionsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CategoryPermissionsAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	cid, _ := strconv.Atoi(vars["category"])
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.PermissionUserAllow(r.Context(), db.PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section:      sql.NullString{Valid: true, String: fmt.Sprintf("writing:%d", cid)},
		Role:         sql.NullString{Valid: true, String: level},
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func CategoryPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	permid, _ := strconv.Atoi(r.PostFormValue("permid"))
	if err := queries.PermissionUserDisallow(r.Context(), int32(permid)); err != nil {
		log.Printf("permissionUserDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
