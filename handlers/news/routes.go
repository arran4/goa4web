package news

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/gorilla/mux"

	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

func AddNewsIndex(h http.Handler) http.Handler { return handlers.IndexMiddleware(CustomNewsIndex)(h) }

func runTemplate(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlers.TemplateHandler(w, r, name, r.Context().Value(common.KeyCoreData))
	}
}

// RegisterRoutes attaches the public news endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("News", "/", SectionWeight)
	nav.RegisterAdminControlCenter("News", "/admin/news/users/levels", SectionWeight)
	r.Use(handlers.IndexMiddleware(CustomNewsIndex))
	r.HandleFunc("/", runTemplate("newsPage.gohtml")).Methods("GET")
	r.HandleFunc("/", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	r.HandleFunc("/news.rss", NewsRssPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(handlers.IndexMiddleware(CustomNewsIndex))
	nr.HandleFunc("", runTemplate("newsPage.gohtml")).Methods("GET")
	nr.HandleFunc("", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", replyTask.Action).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Match)
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(NewsPostCommentEditActionPage))).Methods("POST").MatcherFunc(tasks.EditReplyTask.Match)
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(NewsPostCommentEditActionCancelPage))).Methods("POST").MatcherFunc(tasks.CancelTask.Match)
	nr.HandleFunc("/news/{post}", editTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editTask.Match)
	nr.HandleFunc("/news/{post}", newPostTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(newPostTask.Match)
	nr.HandleFunc("/news/{post}/announcement", announcementAddTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementAddTask.Match)
	nr.HandleFunc("/news/{post}/announcement", announcementDeleteTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementDeleteTask.Match)
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(tasks.CancelTask.Match)
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", NewsUserPermissionsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", userAllowTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Match)
	nr.HandleFunc("/users/permissions", userDisallowTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Match)
}

// Register registers the news router module.
func Register() {
	router.RegisterModule("news", nil, RegisterRoutes)
}
