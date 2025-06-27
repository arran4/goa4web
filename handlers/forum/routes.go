package forum

import (
	"github.com/gorilla/mux"
	"net/http"

	comments "github.com/arran4/goa4web/handlers/comments"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterRoutes attaches the public forum endpoints to r.
func RegisterRoutes(r *mux.Router) {
	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("/topic/{topic}.rss", TopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", TopicAtomPage).Methods("GET")
	fr.HandleFunc("", Page).Methods("GET")
	fr.HandleFunc("/category/{category}", Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}", TopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", ThreadNewPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", ThreadNewActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCreateThread))
	fr.HandleFunc("/topic/{topic}/thread", ThreadNewCancelPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods("GET")
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(hcommon.TaskDoneAutoRefreshPage))).Methods("POST")
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(TopicThreadReplyPage))).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(TopicThreadReplyCancelPage))).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(TopicThreadCommentEditActionPage)))).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEditReply))
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(http.HandlerFunc(TopicThreadCommentEditActionCancelPage))).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
}
