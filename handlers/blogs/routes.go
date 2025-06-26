package blogs

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	comments "github.com/arran4/goa4web/handlers/comments"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterRoutes attaches the public blog endpoints to r.
func RegisterRoutes(r *mux.Router) {
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", RssPage).Methods("GET")
	br.HandleFunc("/atom", AtomPage).Methods("GET")
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/", Page).Methods("GET")
	br.HandleFunc("/add", BlogAddPage).Methods("GET").MatcherFunc(auth.RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", BlogAddActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskAdd))
	br.HandleFunc("/bloggers", BloggersPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", BloggerPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", CommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", BlogReplyPostPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", CommentEditPostPage).MatcherFunc(Or(auth.RequiredAccess("administrator"), comments.Author())).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEditReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", CommentEditPostCancelPage).MatcherFunc(Or(auth.RequiredAccess("administrator"), comments.Author())).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	br.HandleFunc("/blog/{blog}/edit", BlogEditPage).Methods("GET").MatcherFunc(Or(auth.RequiredAccess("administrator"), And(auth.RequiredAccess("writer"), BlogAuthor())))
	br.HandleFunc("/blog/{blog}/edit", BlogEditActionPage).Methods("POST").MatcherFunc(Or(auth.RequiredAccess("administrator"), And(auth.RequiredAccess("writer"), BlogAuthor()))).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEdit))
	br.HandleFunc("/blog/{blog}/edit", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserAllow))
	br.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserDisallow))
	br.HandleFunc("/users/permissions", UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUsersAllow))
	br.HandleFunc("/users/permissions", UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUsersDisallow))
}
