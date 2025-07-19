package imagebbs

import (
	"github.com/gorilla/mux"

	handlers "github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	ar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", AdminBoardsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board", AdminNewBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board", newBoardTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(newBoardTask.Matcher())
	ar.HandleFunc("/board", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board/{board}", modifyBoardTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(modifyBoardTask.Matcher())
	ar.HandleFunc("/approve/{post}", approvePostTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(approvePostTask.Matcher())
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
}
