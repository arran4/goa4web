package forum

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/handlers"
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
	fr.HandleFunc("/topic/{topic}/subscribe", tasks.Action(subscribeTopicTaskAction)).Methods("POST").MatcherFunc(subscribeTopicTaskAction.Matcher())
	fr.HandleFunc("/topic/{topic}/unsubscribe", tasks.Action(unsubscribeTopicTaskAction)).Methods("POST").MatcherFunc(unsubscribeTopicTaskAction.Matcher())
	fr.HandleFunc("/topic/{topic}/thread", createThreadTask.Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", tasks.Action(createThreadTask)).Methods("POST").MatcherFunc(createThreadTask.Matcher())
	fr.HandleFunc("/topic/{topic}/thread/cancel", ThreadNewCancelPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread/cancel", tasks.Action(threadNewCancelAction)).Methods("POST").MatcherFunc(threadNewCancelAction.Matcher())
	fr.HandleFunc("/topic/{topic}/thread", tasks.Action(threadNewCancelAction)).Methods("POST").MatcherFunc(threadNewCancelAction.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods("GET")
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods("POST")
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(tasks.Action(replyTask)))).Methods("POST").MatcherFunc(replyTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(tasks.Action(topicThreadReplyCancel)))).Methods("POST").MatcherFunc(topicThreadReplyCancel.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(tasks.Action(topicThreadCommentEditAction))))).Methods("POST").MatcherFunc(topicThreadCommentEditAction.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(tasks.Action(topicThreadCommentEditActionCancel))))).Methods("POST").MatcherFunc(topicThreadCommentEditActionCancel.Matcher())
}

// Register registers the forum router module.
func Register() {
	router.RegisterModule("forum", nil, RegisterRoutes)
}
