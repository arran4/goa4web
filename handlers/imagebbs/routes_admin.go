package imagebbs

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	iar := ar.PathPrefix("/imagebbs").Subrouter()
	iar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	iar.HandleFunc("/boards", AdminBoardsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	iar.HandleFunc("/board", AdminNewBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	iar.HandleFunc("/board", handlers.TaskHandler(newBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(newBoardTask.Matcher())
	iar.HandleFunc("/board/{board}", handlers.TaskHandler(modifyBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(modifyBoardTask.Matcher())
	iar.HandleFunc("/approve/{post}", handlers.TaskHandler(approvePostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(approvePostTask.Matcher())
	iar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
}
