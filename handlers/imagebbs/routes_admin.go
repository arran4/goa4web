package imagebbs

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	ar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", AdminBoardsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board", AdminNewBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board", tasks.Action(newBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(newBoardTask.Matcher())
	ar.HandleFunc("/board", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/board/{board}", tasks.Action(modifyBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(modifyBoardTask.Matcher())
	ar.HandleFunc("/approve/{post}", tasks.Action(approvePostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(approvePostTask.Matcher())
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
}
