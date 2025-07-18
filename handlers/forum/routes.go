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
	fr.HandleFunc("/topic/{topic}/thread", threadNewCancel.Action).Methods("POST").MatcherFunc(threadNewCancelPage.Matcher()) // TODO make this form
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods("GET")
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods("POST")
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(replyTask.Action))).Methods("POST").MatcherFunc(replyTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(topicThreadReplyCancel.Action))).Methods("POST").MatcherFunc(topicThreadReplyCancelPage.Matcher())                                                        // TODO make this form
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(topicThreadCommentEditAction.Action)))).Methods("POST").MatcherFunc(topicThreadCommentEditActionPage.Matcher()) // TODO make this form
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(http.HandlerFunc(topicThreadCommentEditActionCancel.Action))).Methods("POST").MatcherFunc(topicThreadCommentEditActionCancelPage.Matcher())                    // TODO make this form
}

// Register registers the forum router module.
func Register() {
	router.RegisterModule("forum", nil, RegisterRoutes)
}
