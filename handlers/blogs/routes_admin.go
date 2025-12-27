package blogs

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches blog admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	br := ar.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequireAdminAccess())
	br.HandleFunc("/", AdminPage).Methods("GET").MatcherFunc(handlers.RequireAdminAccess())
	br.HandleFunc("/blog/{blog}", AdminBlogPage).Methods("GET").MatcherFunc(handlers.RequireAdminAccess())
	br.HandleFunc("/blog/{blog}/edit", AdminBlogEditPage).Methods("GET").MatcherFunc(handlers.RequireAdminAccess())
	br.HandleFunc("/blog/{blog}/comments", AdminBlogCommentsPage).Methods("GET").MatcherFunc(handlers.RequireAdminAccess())
}
