package news

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/gorilla/mux"

	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

func AddNewsIndex(h http.Handler) http.Handler { return handlers.IndexMiddleware(CustomNewsIndex)(h) }

func runTemplate(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlers.TemplateHandler(w, r, name, r.Context().Value(handlers.KeyCoreData))
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
	nr.HandleFunc("/news/{post}", ReplyTask.Action).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(ReplyTask.Match)
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(NewsPostCommentEditActionPage))).Methods("POST").MatcherFunc(tasks.EditReplyTask.Match)
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(NewsPostCommentEditActionCancelPage))).Methods("POST").MatcherFunc(tasks.CancelTask.Match)
	nr.HandleFunc("/news/{post}", EditTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(EditTask.Match)
	nr.HandleFunc("/news/{post}", NewPostTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(NewPostTask.Match)
	nr.HandleFunc("/news/{post}/announcement", AnnouncementAddTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(AnnouncementAddTask.Match)
	nr.HandleFunc("/news/{post}/announcement", AnnouncementDeleteTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(AnnouncementDeleteTask.Match)
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(tasks.CancelTask.Match)
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", NewsUserPermissionsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", UserAllowTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(UserAllowTask.Match)
	nr.HandleFunc("/users/permissions", UserDisallowTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(UserDisallowTask.Match)
}

// Register registers the news router module.
func Register() {
	router.RegisterModule("news", nil, RegisterRoutes)
}
