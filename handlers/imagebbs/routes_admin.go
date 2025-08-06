package imagebbs

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	iar := ar.PathPrefix("/imagebbs").Subrouter()
	iar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br := iar.PathPrefix("/boards").Subrouter()
	br.HandleFunc("", AdminBoardsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/new", AdminNewBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/new", handlers.TaskHandler(newBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(newBoardTask.Matcher())
	bb := br.PathPrefix("/board/{board}").Subrouter()
	bb.HandleFunc("", AdminBoardDashboardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	bb.HandleFunc("/edit", AdminBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	bb.HandleFunc("/edit", handlers.TaskHandler(modifyBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(modifyBoardTask.Matcher())
	bb.HandleFunc("/list", AdminBoardListPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	bb.HandleFunc("/delete", handlers.TaskHandler(deleteBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(deleteBoardTask.Matcher())
	bb.HandleFunc("/images", AdminBoardImagesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	iar.HandleFunc("/approve/{post}", handlers.TaskHandler(approvePostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(approvePostTask.Matcher())
}
