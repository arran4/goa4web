package imagebbs

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches image board admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	iar := ar.PathPrefix("/imagebbs").Subrouter()
	iar.HandleFunc("", AdminPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	br := iar.PathPrefix("/boards").Subrouter()
	br.HandleFunc("", AdminBoardsPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	br.HandleFunc("/new", AdminNewBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	br.HandleFunc("/new", handlers.TaskHandler(newBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(newBoardTask.Matcher())

	bb := iar.PathPrefix("/board/{board}").Subrouter()
	bb.HandleFunc("", AdminBoardViewPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	bb.HandleFunc("/edit", AdminBoardPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	bb.HandleFunc("/edit", handlers.TaskHandler(modifyBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(modifyBoardTask.Matcher())
	bb.HandleFunc("/images", AdminBoardListPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	bb.HandleFunc("/delete", handlers.TaskHandler(deleteBoardTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(deleteBoardTask.Matcher())
	iar.HandleFunc("/approve/{post}", handlers.TaskHandler(approvePostTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(approvePostTask.Matcher())

	uar := ar.PathPrefix("/user/{user}/imagebbs").Subrouter()
	uar.HandleFunc("/post/{post}", AdminPostDashboardPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	uar.HandleFunc("/post/{post}/edit", AdminPostEditPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
	uar.HandleFunc("/post/{post}/edit", handlers.TaskHandler(modifyPostTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(modifyPostTask.Matcher())
	uar.HandleFunc("/post/{post}/edit", handlers.TaskHandler(deletePostTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(deletePostTask.Matcher())
	uar.HandleFunc("/post/{post}/comments", AdminPostCommentsPage).Methods("GET").MatcherFunc(handlers.RequiredAdminAccess())
}
