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
	ar.HandleFunc("/board", NewBoardTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(NewBoardTask.Match)
	ar.HandleFunc("/board", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board/{board}", ModifyBoardTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(ModifyBoardTask.Match)
	ar.HandleFunc("/approve/{post}", ApprovePostTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(ApprovePostTask.Match)
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
}
