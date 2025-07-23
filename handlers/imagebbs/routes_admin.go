package imagebbs

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	ar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", AdminBoardsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board", AdminNewBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board", handlers.TaskHandler(newBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(newBoardTask.Matcher())
	ar.HandleFunc("/board/{board}", handlers.TaskHandler(modifyBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(modifyBoardTask.Matcher())
	ar.HandleFunc("/approve/{post}", handlers.TaskHandler(approvePostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(approvePostTask.Matcher())
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
}
