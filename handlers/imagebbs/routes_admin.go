package imagebbs

import (
	"github.com/gorilla/mux"

	handlers "github.com/arran4/goa4web/handlers"
	auth "github.com/arran4/goa4web/handlers/auth"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	ar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", AdminBoardsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board", AdminNewBoardPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board", newBoardTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(newBoardTask.Match)
	ar.HandleFunc("/board", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board/{board}", modifyBoardTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(modifyBoardTask.Match)
	ar.HandleFunc("/approve/{post}", approvePostTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(approvePostTask.Match)
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
}
