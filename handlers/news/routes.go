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
	navReg.RegisterIndexLink("News", "/", SectionWeight)
	navReg.RegisterAdminControlCenter("News", "News", "/admin/news", SectionWeight)
	r.Use(handlers.IndexMiddleware(CustomNewsIndex))
	r.HandleFunc("/", NewsPage).Methods("GET")
	r.HandleFunc("/", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	r.HandleFunc("/news.rss", NewsRssPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(handlers.IndexMiddleware(CustomNewsIndex))
	nr.HandleFunc("", NewsPage).Methods("GET")
	nr.HandleFunc("", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{news}", NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	nr.Handle("/news/{news}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	nr.Handle("/news/{news}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(editTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskHandler(newPostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(newPostTask.Matcher())
	nr.HandleFunc("/news/{news}/announcement", handlers.TaskHandler(announcementAddTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementAddTask.Matcher())
	nr.HandleFunc("/news/{news}/announcement", handlers.TaskHandler(announcementDeleteTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementDeleteTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{news}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
}

// Register registers the news router module.
func Register(reg *router.Registry) {
	log.Printf("News: Registering")
	reg.RegisterModule("news", nil, RegisterRoutes)
}
