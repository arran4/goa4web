package blogs

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the public blog endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Blogs", "/blogs", SectionWeight)
	navReg.RegisterAdminControlCenter("Blogs", "Blogs", "/admin/blogs", SectionWeight)
	br := r.PathPrefix("/blogs").Subrouter()
	br.Use(handlers.IndexMiddleware(CustomBlogIndex))
	br.HandleFunc("/rss", RssPage).Methods("GET")
	br.HandleFunc("/atom", AtomPage).Methods("GET")
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/", Page).Methods("GET")
	br.HandleFunc("/add", addBlogTask.Page).Methods("GET").MatcherFunc(handlers.RequiredAccess("content writer", "administrator"))
	br.HandleFunc("/add", handlers.TaskHandler(addBlogTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(addBlogTask.Matcher())
	br.HandleFunc("/bloggers", BloggerListPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", CommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", handlers.TaskHandler(replyBlogTask)).Methods("POST").MatcherFunc(replyBlogTask.Matcher())
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(editBlogTask.Page))).Methods("GET").MatcherFunc(handlers.RequiredAccess("content writer", "administrator"))
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(handlers.TaskHandler(editBlogTask)))).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editBlogTask.Matcher())
	br.HandleFunc("/blog/{blog}/edit", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Matcher())
	br.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Matcher())
	br.HandleFunc("/users/permissions", UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(usersAllowTask.Matcher())
	br.HandleFunc("/users/permissions", UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(usersDisallowTask.Matcher())
}

// Register registers the blogs router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("blogs", nil, RegisterRoutes)
}
