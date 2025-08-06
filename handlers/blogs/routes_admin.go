package blogs

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches blog admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	br := ar.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/users/roles", GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/users/roles", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Matcher())
	br.HandleFunc("/users/roles", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Matcher())
	br.HandleFunc("/users/roles", UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(usersAllowTask.Matcher())
	br.HandleFunc("/users/roles", UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(usersDisallowTask.Matcher())
	br.HandleFunc("/blog/{blog}", AdminBlogPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/blog/{blog}/edit", AdminBlogEditPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/blog/{blog}/comments", AdminBlogCommentsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
}
