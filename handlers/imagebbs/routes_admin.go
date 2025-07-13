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
	ar.HandleFunc("/board", AdminNewBoardMakePage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(NewBoardTask.Matcher)
	ar.HandleFunc("/board", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ar.HandleFunc("/board/{board}", AdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(ModifyBoardTask.Matcher)
	ar.HandleFunc("/approve/{post}", AdminApprovePostPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(ApprovePostTask.Matcher)
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
}
