package writings

import (
	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserAllow))
	war.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserDisallow))
	war.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserAllow))
	war.HandleFunc("/users/levels", AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserDisallow))
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", AdminCategoriesModifyPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskWritingCategoryChange))
	war.HandleFunc("/categories", AdminCategoriesCreatePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskWritingCategoryCreate))
}
