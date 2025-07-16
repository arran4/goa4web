package blogs

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
	"net/http"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// AddBlogIndex injects blog index links into CoreData.
func AddBlogIndex(h http.Handler) http.Handler { return hcommon.IndexMiddleware(CustomBlogIndex)(h) }

// RegisterRoutes attaches the public blog endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Blogs", "/blogs", SectionWeight)
	nav.RegisterAdminControlCenter("Blogs", "/admin/blogs/user/permissions", SectionWeight)
	br := r.PathPrefix("/blogs").Subrouter()
	br.Use(hcommon.IndexMiddleware(CustomBlogIndex))
	br.HandleFunc("/rss", RssPage).Methods("GET")
	br.HandleFunc("/atom", AtomPage).Methods("GET")
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/", Page).Methods("GET")
	br.HandleFunc("/add", AddBlogTask.Page).Methods("GET").MatcherFunc(auth.RequiredAccess("content writer", "administrator"))
	br.HandleFunc("/add", AddBlogTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("content writer", "administrator")).MatcherFunc(AddBlogTask.Match)
	br.HandleFunc("/bloggers", BloggerListPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", CommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", ReplyBlogTask.Action).Methods("POST").MatcherFunc(ReplyBlogTask.Match)
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditPostPage))).Methods("POST").MatcherFunc(tasks.EditReplyTask.Match)
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditPostCancelPage))).Methods("POST").MatcherFunc(tasks.CancelTask.Match)
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(EditBlogTask.Page))).Methods("GET").MatcherFunc(auth.RequiredAccess("content writer", "administrator"))
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(EditBlogTask.Action))).Methods("POST").MatcherFunc(auth.RequiredAccess("content writer", "administrator")).MatcherFunc(EditBlogTask.Match)
	br.HandleFunc("/blog/{blog}/edit", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(tasks.CancelTask.Match)

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UserAllowTask.Match)
	br.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UserDisallowTask.Match)
	br.HandleFunc("/users/permissions", UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UsersAllowTask.Match)
	br.HandleFunc("/users/permissions", UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UsersDisallowTask.Match)
}

// Register registers the blogs router module.
func Register() {
	router.RegisterModule("blogs", nil, RegisterRoutes)
}
