package news

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches news admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	nr := ar.PathPrefix("/news").Subrouter()
	nr.HandleFunc("", AdminNewsPage).Methods("GET")
	nr.HandleFunc("/{post}", AdminNewsPostPage).Methods("GET")
	nr.HandleFunc("/{post}/edit", adminNewsEditFormPage).Methods("GET")
	nr.HandleFunc("/{post}/edit", handlers.TaskHandler(editTask)).Methods("POST").MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/{post}/delete", AdminNewsDeleteConfirmPage).Methods("GET")
	nr.HandleFunc("/{post}/delete", handlers.TaskHandler(deleteNewsPostTask)).Methods("POST").MatcherFunc(deleteNewsPostTask.Matcher())
}
