package writings

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterRoutes attaches the public writings endpoints to r.
func RegisterRoutes(r *mux.Router) {
	wr := r.PathPrefix("/writings").Subrouter()
	wr.HandleFunc("/rss", RssPage).Methods("GET")
	wr.HandleFunc("/atom", AtomPage).Methods("GET")
	wr.HandleFunc("", Page).Methods("GET")
	wr.HandleFunc("/", Page).Methods("GET")
	wr.HandleFunc("/writer/{username}", WriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", WriterPage).Methods("GET")
	wr.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserAllow))
	wr.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserDisallow))
	wr.HandleFunc("/article/{article}", ArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", ArticleReplyActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	wr.HandleFunc("/article/{article}/edit", ArticleEditPage).Methods("GET").MatcherFunc(Or(And(auth.RequiredAccess("writer"), WritingAuthor()), auth.RequiredAccess("administrator")))
	wr.HandleFunc("/article/{article}/edit", ArticleEditActionPage).Methods("POST").MatcherFunc(Or(And(auth.RequiredAccess("writer"), WritingAuthor()), auth.RequiredAccess("administrator"))).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUpdateWriting))
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", ArticleAddPage).Methods("GET").MatcherFunc(Or(auth.RequiredAccess("writer"), auth.RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", ArticleAddActionPage).Methods("POST").MatcherFunc(Or(auth.RequiredAccess("writer"), auth.RequiredAccess("administrator"))).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSubmitWriting))
}
