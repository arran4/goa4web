package news

import (
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

func runTemplate(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlers.TemplateHandler(w, r, name, r.Context().Value(consts.KeyCoreData))
	}
}

// RegisterRoutes attaches the public news endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("News", "/", SectionWeight)
	nav.RegisterAdminControlCenter("News", "/admin/news/users/roles", SectionWeight)
	r.Use(handlers.IndexMiddleware(CustomNewsIndex))
	r.HandleFunc("/", runTemplate("newsPage.gohtml")).Methods("GET")
	r.HandleFunc("/", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	r.HandleFunc("/news.rss", NewsRssPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(handlers.IndexMiddleware(CustomNewsIndex))
	nr.HandleFunc("", runTemplate("newsPage.gohtml")).Methods("GET")
	nr.HandleFunc("", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", tasks.Action(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(tasks.Action(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(tasks.Action(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{post}", tasks.Action(editTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{post}", tasks.Action(newPostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(newPostTask.Matcher())
	nr.HandleFunc("/news/{post}/announcement", tasks.Action(announcementAddTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementAddTask.Matcher())
	nr.HandleFunc("/news/{post}/announcement", tasks.Action(announcementDeleteTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementDeleteTask.Matcher())
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", NewsUserPermissionsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", tasks.Action(userAllowTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Matcher())
	nr.HandleFunc("/users/permissions", tasks.Action(userDisallowTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Matcher())
}

// Register registers the news router module.
func Register() {
	router.RegisterModule("news", nil, RegisterRoutes)
}
