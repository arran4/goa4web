package blogs

import (
	"github.com/gorilla/mux"
	"net/http"

	auth "github.com/arran4/goa4web/handlers/auth"
	comments "github.com/arran4/goa4web/handlers/comments"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the public blog endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Blogs", "/blogs", SectionWeight)
	nav.RegisterAdminControlCenter("Blogs", "/admin/blogs/user/permissions", SectionWeight)
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", RssPage).Methods("GET")
	br.HandleFunc("/atom", AtomPage).Methods("GET")
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/", Page).Methods("GET")
	br.HandleFunc("/add", BlogAddPage).Methods("GET").MatcherFunc(auth.RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", BlogAddActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(AddBlogTask.Matcher)
	br.HandleFunc("/bloggers", BloggerListPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", CommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", BlogReplyPostPage).Methods("POST").MatcherFunc(ReplyBlogTask.Matcher)
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditPostPage))).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEditReply))
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditPostCancelPage))).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(BlogEditPage))).Methods("GET").MatcherFunc(auth.RequiredAccess("writer", "administrator"))
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(BlogEditActionPage))).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(EditBlogTask.Matcher)
	br.HandleFunc("/blog/{blog}/edit", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UserAllowTask.Matcher)
	br.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UserDisallowTask.Matcher)
	br.HandleFunc("/users/permissions", UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UsersAllowTask.Matcher)
	br.HandleFunc("/users/permissions", UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UsersDisallowTask.Matcher)
}

// Register registers the blogs router module.
func Register() {
	router.RegisterModule("blogs", nil, RegisterRoutes)
}
