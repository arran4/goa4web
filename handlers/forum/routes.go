package forum

import (
	auth "github.com/arran4/goa4web/handlers/auth"
	comments "github.com/arran4/goa4web/handlers/comments"
	hcommon "github.com/arran4/goa4web/handlers/common"
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
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
	fr.HandleFunc("/topic/{topic}/thread/{thread}", ThreadPage).Methods("GET").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", TopicThreadReplyPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", TopicThreadReplyCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", TopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEditReply)).MatcherFunc(Or(auth.RequiredAccess("administrator"), comments.Author()))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", TopicThreadCommentEditActionCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
}
