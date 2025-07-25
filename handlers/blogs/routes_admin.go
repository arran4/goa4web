package blogs

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches blog admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	br := ar.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/user/permissions", GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Matcher())
	br.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Matcher())
	br.HandleFunc("/users/permissions", UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(usersAllowTask.Matcher())
	br.HandleFunc("/users/permissions", UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(usersDisallowTask.Matcher())
}
