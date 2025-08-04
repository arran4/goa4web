package news

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches news admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	nr := ar.PathPrefix("/news").Subrouter()
	nr.HandleFunc("", handlers.VerifyAccess(AdminNewsPage, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}", handlers.VerifyAccess(AdminNewsPostPage, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.VerifyAccess(adminNewsEditFormPage, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/edit", handlers.VerifyAccess(handlers.TaskHandler(editTask), "administrator")).Methods("POST").MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/{news}/delete", handlers.VerifyAccess(AdminNewsDeleteConfirmPage, "administrator")).Methods("GET")
	nr.HandleFunc("/{news}/delete", handlers.VerifyAccess(handlers.TaskHandler(deleteNewsPostTask), "administrator")).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())
}
