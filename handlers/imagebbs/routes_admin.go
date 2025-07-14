package imagebbs

import (
	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	ar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", AdminBoardsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/boards", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board", AdminNewBoardPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board", NewBoardTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(NewBoardTask.Match)
	ar.HandleFunc("/board", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board/{board}", ModifyBoardTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(ModifyBoardTask.Match)
	ar.HandleFunc("/approve/{post}", ApprovePostTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(ApprovePostTask.Match)
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
}
