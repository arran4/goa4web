package writings

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(UserAllowTask.Match)
	war.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(UserDisallowTask.Match)
	war.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(UserAllowTask.Match)
	war.HandleFunc("/users/levels", AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(UserDisallowTask.Match)
	war.HandleFunc("/users/access", AdminUserAccessPage).Methods("GET")
	war.HandleFunc("/users/access", AdminUserAccessAddActionPage).Methods("POST").MatcherFunc(AddApprovalTask.Match)
	war.HandleFunc("/users/access", AdminUserAccessUpdateActionPage).Methods("POST").MatcherFunc(UpdateApprovalTask.Match)
	war.HandleFunc("/users/access", AdminUserAccessRemoveActionPage).Methods("POST").MatcherFunc(DeleteApprovalTask.Match)
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", AdminCategoriesModifyPage).Methods("POST").MatcherFunc(WritingCategoryChangeTask.Match)
	war.HandleFunc("/categories", AdminCategoriesCreatePage).Methods("POST").MatcherFunc(WritingCategoryCreateTask.Match)
}
