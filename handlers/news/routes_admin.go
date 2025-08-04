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
	nr.HandleFunc("", handlers.VerifyAccess(AdminNewsPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}", handlers.VerifyAccess(AdminNewsPostPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.VerifyAccess(adminNewsEditFormPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.VerifyAccess(handlers.TaskHandler(editTask), msg, "administrator")).Methods("POST").MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/{news}/delete", handlers.VerifyAccess(AdminNewsDeleteConfirmPage, msg, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/delete", handlers.VerifyAccess(handlers.TaskHandler(deleteNewsPostTask), msg, "administrator")).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())
}
