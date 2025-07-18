package forum

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// AddForumIndex injects forum index links into CoreData.
func AddForumIndex(h http.Handler) http.Handler { return handlers.IndexMiddleware(CustomForumIndex)(h) }

// RegisterRoutes attaches the public forum endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Forum", "/forum", SectionWeight)
	nav.RegisterAdminControlCenter("Forum", "/admin/forum", SectionWeight)
	fr := r.PathPrefix("/forum").Subrouter()
	fr.Use(handlers.IndexMiddleware(CustomForumIndex))
	fr.HandleFunc("/topic/{topic}.rss", TopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", TopicAtomPage).Methods("GET")
	fr.HandleFunc("", Page).Methods("GET")
	fr.HandleFunc("/category/{category}", Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}", TopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", createThreadTask.Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", createThreadTask.Action).Methods("POST").MatcherFunc(createThreadTask.Matcher())
	fr.HandleFunc("/topic/{topic}/thread", ThreadNewCancelPage).Methods("POST").MatcherFunc(tasks.CancelTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods("GET")
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods("POST")
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(replyTask.Action))).Methods("POST").MatcherFunc(replyTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(TopicThreadReplyCancelPage))).Methods("POST").MatcherFunc(tasks.CancelTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(TopicThreadCommentEditActionPage)))).Methods("POST").MatcherFunc(tasks.EditReplyTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(http.HandlerFunc(TopicThreadCommentEditActionCancelPage))).Methods("POST").MatcherFunc(tasks.CancelTask.Matcher())
}

// Register registers the forum router module.
func Register() {
	router.RegisterModule("forum", nil, RegisterRoutes)
}
