package news

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

// RegisterRoutes attaches the public news endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []navpkg.RouterOptions {
	opts := []navpkg.RouterOptions{
		navpkg.NewIndexLinkWithViewPermission("News", "/", SectionWeight, "news", "post"),
		navpkg.NewAdminControlCenterLink(navpkg.AdminCCCategory("News"), "News", "/admin/news", SectionWeight),
	}
	r.Use(handlers.IndexMiddleware(CustomNewsIndex), handlers.SectionMiddleware("news"))
	r.HandleFunc("/", NewsPageHandler).Methods("GET")
	r.HandleFunc("/", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr := r.PathPrefix("/news").Subrouter()
	nr.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	nr.HandleFunc("/rss", NewsRssPage).Methods("GET")
	nr.HandleFunc("/u/{username}/rss", NewsRssPage).Methods("GET")
	nr.Use(handlers.IndexMiddleware(CustomNewsIndex), handlers.SectionMiddleware("news"))
	nr.HandleFunc("", NewsPageHandler).Methods("GET")
	nr.HandleFunc("", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	editGrant := handlers.RequireGrantForPathInt("news", "post", "edit", "news")
	promoteAnnouncementGrant := handlers.RequireGrantForPathInt("news", "post", "promote", "news")
	demoteAnnouncementGrant := handlers.RequireGrantForPathInt("news", "post", "demote", "news")
	nr.HandleFunc("/post", NewsCreatePageHandler).Methods("GET").MatcherFunc(MatchCanPostNews)
	nr.HandleFunc("/post", handlers.TaskHandler(newPostTask)).Methods("POST").MatcherFunc(MatchCanPostNews).MatcherFunc(newPostTask.Matcher())
	nr.HandleFunc("/preview", PreviewPage).Methods("POST")

	// OpenGraph preview endpoint (no auth required for social media bots)
	nr.HandleFunc("/shared/news/{news}", SharedPreviewPage).Methods("GET", "HEAD")
	nr.HandleFunc("/shared/news/{news}/ts/{ts}/sign/{sign}", SharedPreviewPage).Methods("GET", "HEAD")
	nr.HandleFunc("/shared/news/{news}/nonce/{nonce}/sign/{sign}", SharedPreviewPage).Methods("GET", "HEAD")

	nr.HandleFunc("/news/{news}", NewsPostPageHandler).Methods("GET")
	nr.Handle("/news/{news}/edit", RequireNewsPostAuthor(http.HandlerFunc(editTask.Page))).Methods("GET").MatcherFunc(editGrant)
	nr.Handle("/news/{news}/edit", RequireNewsPostAuthor(http.HandlerFunc(handlers.TaskHandler(editTask)))).Methods("POST").MatcherFunc(editGrant).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	nr.Handle("/news/{news}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	nr.Handle("/news/{news}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(editTask)).Methods("POST").MatcherFunc(editGrant).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{news}/labels", handlers.TaskHandler(setLabelsTask)).Methods("POST").MatcherFunc(setLabelsTask.Matcher())
	nr.HandleFunc("/news/{news}/labels", handlers.TaskHandler(markReadTask)).Methods("GET", "POST").MatcherFunc(markReadTask.Matcher())
	nr.HandleFunc("/news/{news}/announcement", handlers.TaskHandler(announcementAddTask)).Methods("POST").MatcherFunc(promoteAnnouncementGrant).MatcherFunc(announcementAddTask.Matcher())
	nr.HandleFunc("/news/{news}/announcement", handlers.TaskHandler(announcementDeleteTask)).Methods("POST").MatcherFunc(demoteAnnouncementGrant).MatcherFunc(announcementDeleteTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/{news}/labels", handlers.TaskHandler(markReadTask)).Methods("GET").MatcherFunc(markReadTask.Matcher())

	api := r.PathPrefix("/api/news").Subrouter()
	api.HandleFunc("/share", share.ShareLink).Methods("GET")
	return opts
}

// Register registers the news router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("news", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []navpkg.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
