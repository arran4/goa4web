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
	nr.HandleFunc("", handlers.RequireRole(AdminNewsPage, msg, "administrator")).Methods("GET")

	// Article routes
	articleRouter := nr.PathPrefix("/article").Subrouter()
	articleRouter.HandleFunc("/{news}", handlers.RequireRole(AdminNewsPostPage, msg, "administrator")).Methods("GET")
	articleRouter.HandleFunc("/{news}/edit", handlers.RequireRole(adminNewsEditFormPage, msg, "administrator")).Methods("GET")
	articleRouter.HandleFunc("/{news}/edit", handlers.RequireRole(handlers.TaskHandler(editTask), msg, "administrator")).Methods("POST").MatcherFunc(editTask.Matcher())
	articleRouter.HandleFunc("/{news}/delete", handlers.RequireRole(AdminNewsDeleteConfirmPage, msg, "administrator")).Methods("GET")
	articleRouter.HandleFunc("/{news}/delete", handlers.RequireRole(handlers.TaskHandler(deleteNewsPostTask), msg, "administrator")).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())

	// Legacy routes for compatibility
	nr.HandleFunc("/{news}", handlers.RequireRole(AdminNewsPostPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.RequireRole(adminNewsEditFormPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.RequireRole(handlers.TaskHandler(editTask), msg, "administrator")).Methods("POST").MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/{news}/delete", handlers.RequireRole(AdminNewsDeleteConfirmPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/delete", handlers.RequireRole(handlers.TaskHandler(deleteNewsPostTask), msg, "administrator")).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())
}
