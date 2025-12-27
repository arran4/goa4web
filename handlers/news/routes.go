package news

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/handlers/forum/comments"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the public news endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	log.Printf("News: Registering Routes")
	navReg.RegisterIndexLinkWithViewPermission("News", "/", SectionWeight, "news", "post")
	navReg.RegisterAdminControlCenter("News", "News", "/admin/news", SectionWeight)
	r.Use(handlers.IndexMiddleware(CustomNewsIndex), handlers.SectionMiddleware("news"))
	r.HandleFunc("/", NewsPageHandler).Methods("GET")
	r.HandleFunc("/", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr := r.PathPrefix("/news").Subrouter()
	nr.HandleFunc("/rss", NewsRssPage).Methods("GET")
	nr.HandleFunc("/u/{username}/rss", NewsRssPage).Methods("GET")
	nr.Use(handlers.IndexMiddleware(CustomNewsIndex), handlers.SectionMiddleware("news"))
	nr.HandleFunc("", NewsPageHandler).Methods("GET")
	nr.HandleFunc("", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/post", NewsCreatePageHandler).Methods("GET").MatcherFunc(MatchCanPostNews)
	nr.HandleFunc("/post", handlers.TaskHandler(newPostTask)).Methods("POST").MatcherFunc(MatchCanPostNews).MatcherFunc(newPostTask.Matcher())
	nr.HandleFunc("/news/{news}", NewsPostPageHandler).Methods("GET")
	nr.Handle("/news/{news}/edit", RequireNewsPostAuthor(http.HandlerFunc(editTask.Page))).Methods("GET").MatcherFunc(handlers.RequiredAccess("content writer", "administrator"))
	nr.Handle("/news/{news}/edit", RequireNewsPostAuthor(http.HandlerFunc(handlers.TaskHandler(editTask)))).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	nr.Handle("/news/{news}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	nr.Handle("/news/{news}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(editTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{news}/labels", handlers.TaskHandler(setLabelsTask)).Methods("POST").MatcherFunc(setLabelsTask.Matcher())
	nr.HandleFunc("/news/{news}/labels", handlers.TaskHandler(markReadTask)).Methods("GET", "POST").MatcherFunc(markReadTask.Matcher())
	nr.HandleFunc("/news/{news}/announcement", handlers.TaskHandler(announcementAddTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(announcementAddTask.Matcher())
	nr.HandleFunc("/news/{news}/announcement", handlers.TaskHandler(announcementDeleteTask)).Methods("POST").MatcherFunc(handlers.RequiredAdminAccess()).MatcherFunc(announcementDeleteTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
}

// Register registers the news router module.
func Register(reg *router.Registry) {
	log.Printf("News: Registering")
	reg.RegisterModule("news", nil, RegisterRoutes)
}
