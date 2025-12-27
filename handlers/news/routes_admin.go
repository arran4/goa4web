package news

import (
	"fmt"

	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches news admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	nr := ar.PathPrefix("/news").Subrouter()
	msg := fmt.Errorf("administrator role required")
	nr.HandleFunc("", handlers.VerifyAdminAccess(AdminNewsPage, msg)).Methods("GET")

	// Article routes
	articleRouter := nr.PathPrefix("/article").Subrouter()
	articleRouter.HandleFunc("/{news}", handlers.VerifyAdminAccess(AdminNewsPostPage, msg)).Methods("GET")
	articleRouter.HandleFunc("/{news}/edit", handlers.VerifyAdminAccess(adminNewsEditFormPage, msg)).Methods("GET")
	articleRouter.HandleFunc("/{news}/edit", handlers.VerifyAdminAccess(handlers.TaskHandler(editTask), msg)).Methods("POST").MatcherFunc(editTask.Matcher())
	articleRouter.HandleFunc("/{news}/delete", handlers.VerifyAdminAccess(AdminNewsDeleteConfirmPage, msg)).Methods("GET")
	articleRouter.HandleFunc("/{news}/delete", handlers.VerifyAdminAccess(handlers.TaskHandler(deleteNewsPostTask), msg)).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())

	// Legacy routes for compatibility
	nr.HandleFunc("/{news}", handlers.VerifyAdminAccess(AdminNewsPostPage, msg)).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.VerifyAdminAccess(adminNewsEditFormPage, msg)).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.VerifyAdminAccess(handlers.TaskHandler(editTask), msg)).Methods("POST").MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/{news}/delete", handlers.VerifyAdminAccess(AdminNewsDeleteConfirmPage, msg)).Methods("GET")
	nr.HandleFunc("/{news}/delete", handlers.VerifyAdminAccess(handlers.TaskHandler(deleteNewsPostTask), msg)).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())
}
