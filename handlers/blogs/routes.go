package blogs

import (
	"net/http"

	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	"github.com/arran4/goa4web/handlers/share"
	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the public blog endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLinkWithViewPermission("Blogs", "/blogs", SectionWeight, "blogs", "entry")
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Blogs"), "Blogs", "/admin/blogs", SectionWeight)
	br := r.PathPrefix("/blogs").Subrouter()
	br.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	br.Use(handlers.IndexMiddleware(BlogsMiddlewareIndex), handlers.SectionMiddleware("blogs"))
	br.HandleFunc("/rss", RssPage).Methods("GET")
	br.HandleFunc("/atom", AtomPage).Methods("GET")
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/", Page).Methods("GET")
	br.HandleFunc("/preview", handlers.PreviewPage).Methods("POST")
	br.HandleFunc("/add", addBlogTask.Page).Methods("GET").MatcherFunc(handlers.RequiredGrant("blogs", "entry", "post", 0))
	br.HandleFunc("/add", handlers.TaskHandler(addBlogTask)).Methods("POST").MatcherFunc(handlers.RequiredGrant("blogs", "entry", "post", 0)).MatcherFunc(addBlogTask.Matcher())
	br.HandleFunc("/bloggers", BloggerListPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPostsPage).Methods("GET")

	// OpenGraph preview endpoint (no auth required for social media bots)
	br.HandleFunc("/shared/blog/{blog}", SharedPreviewPage).Methods("GET", "HEAD")
	br.HandleFunc("/shared/blog/{blog}/ts/{ts}/sign/{sign}", SharedPreviewPage).Methods("GET", "HEAD")
	br.HandleFunc("/shared/blog/{blog}/nonce/{nonce}/sign/{sign}", SharedPreviewPage).Methods("GET", "HEAD")

	br.HandleFunc("/blog/{blog}", BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", BlogsCommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", handlers.TaskHandler(replyBlogTask)).Methods("POST").MatcherFunc(replyBlogTask.Matcher())
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	br.Handle("/blog/{blog}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(editBlogTask.Page))).Methods("GET").MatcherFunc(RequireBlogEditGrant())
	br.Handle("/blog/{blog}/edit", RequireBlogAuthor(http.HandlerFunc(handlers.TaskHandler(editBlogTask)))).Methods("POST").MatcherFunc(RequireBlogEditGrant()).MatcherFunc(editBlogTask.Matcher())
	br.HandleFunc("/blog/{blog}/edit", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())
	br.HandleFunc("/blog/{blog}/labels", handlers.TaskHandler(setLabelsTask)).Methods("POST").MatcherFunc(setLabelsTask.Matcher())
	br.HandleFunc("/blog/{blog}/labels", handlers.TaskHandler(markBlogReadTask)).Methods("GET").MatcherFunc(markBlogReadTask.Matcher())

	api := r.PathPrefix("/api/blogs").Subrouter()
	api.HandleFunc("/share", share.ShareLink).Methods("GET")
}

// Register registers the blogs router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("blogs", nil, RegisterRoutes)
}
