package forum

import (
	"net/http"

	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	"github.com/arran4/goa4web/handlers/share"
	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the public forum endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLinkWithViewPermission("Forum", "/forum", SectionWeight, "forum", "category")
	navReg.RegisterAdminControlCenter("Forum", "Forum", "/admin/forum", SectionWeight)
	fr := r.PathPrefix("/forum").Subrouter()
	fr.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	h := New()
	fr.HandleFunc("/forum.js", h.serveJS).Methods("GET")
	fr.HandleFunc("/forum.css", h.serveCSS).Methods("GET")
	fr.Use(handlers.IndexMiddleware(CustomForumIndex), handlers.SectionMiddleware("forum"))
	fr.HandleFunc("/topic/{topic}.rss", TopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.rss/u/{username}", TopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", TopicAtomPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom/u/{username}", TopicAtomPage).Methods("GET")
	fr.HandleFunc("", Page).Methods("GET")
	fr.HandleFunc("/category/{category}", Page).Methods("GET")
	fr.HandleFunc("/categories/category/{category}", Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}", TopicsPage).Methods("GET")
	// Confirmation pages for subscribe/unsubscribe (GET), posting to task endpoints (POST)
	fr.HandleFunc("/topic/{topic}/subscribe", SubscribeTopicPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/unsubscribe", UnsubscribeTopicPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/subscribe", handlers.TaskHandler(subscribeTopicTaskAction)).Methods("POST").MatcherFunc(subscribeTopicTaskAction.Matcher())
	fr.HandleFunc("/topic/{topic}/unsubscribe", handlers.TaskHandler(unsubscribeTopicTaskAction)).Methods("POST").MatcherFunc(unsubscribeTopicTaskAction.Matcher())
	fr.HandleFunc("/topic_labels.js", handlers.TopicLabelsJS(cfg)).Methods(http.MethodGet)
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(markThreadReadTask)).Methods(http.MethodGet)
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(setLabelsTask)).Methods("POST").MatcherFunc(setLabelsTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(addPublicLabelTask)).Methods("POST").MatcherFunc(addPublicLabelTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(removePublicLabelTask)).Methods("POST").MatcherFunc(removePublicLabelTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(addAuthorLabelTask)).Methods("POST").MatcherFunc(addAuthorLabelTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(removeAuthorLabelTask)).Methods("POST").MatcherFunc(removeAuthorLabelTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(addPrivateLabelTask)).Methods("POST").MatcherFunc(addPrivateLabelTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(removePrivateLabelTask)).Methods("POST").MatcherFunc(removePrivateLabelTask.Matcher())
	fr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(markThreadReadTask)).Methods("POST").MatcherFunc(markThreadReadTask.Matcher())
	fr.HandleFunc("/topic/{topic}/thread", createThreadTask.Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(createThreadTask)).Methods("POST").MatcherFunc(createThreadTask.Matcher())
	fr.HandleFunc("/topic/{topic}/thread/cancel", ThreadNewCancelPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread/cancel", handlers.TaskHandler(threadNewCancelAction)).Methods("POST").MatcherFunc(threadNewCancelAction.Matcher())
	fr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(threadNewCancelAction)).Methods("POST").MatcherFunc(threadNewCancelAction.Matcher())

	// OpenGraph preview endpoint (no auth required for social media bots if signed)
	fr.HandleFunc("/shared/topic/{topic}", SharedTopicPreviewPage).Methods("GET", "HEAD")
	fr.HandleFunc("/shared/topic/{topic}/ts/{ts}/sign/{sign}", SharedTopicPreviewPage).Methods("GET", "HEAD")
	fr.HandleFunc("/shared/topic/{topic}/nonce/{nonce}/sign/{sign}", SharedTopicPreviewPage).Methods("GET", "HEAD")
	fr.HandleFunc("/shared/topic/{topic}/thread/{thread}", SharedThreadPreviewPage).Methods("GET", "HEAD")
	fr.HandleFunc("/shared/topic/{topic}/thread/{thread}/ts/{ts}/sign/{sign}", SharedThreadPreviewPage).Methods("GET", "HEAD")
	fr.HandleFunc("/shared/topic/{topic}/thread/{thread}/nonce/{nonce}/sign/{sign}", SharedThreadPreviewPage).Methods("GET", "HEAD")

	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods("GET")
	fr.Handle("/topic/{topic}/thread/{thread}", RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods("POST")
	fr.Handle("/topic/{topic}/thread/{thread}/reply", RequireThreadAndTopic(http.HandlerFunc(handlers.TaskHandler(replyTask)))).Methods("POST").MatcherFunc(replyTask.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(topicThreadCommentEditAction))))).Methods("POST").MatcherFunc(topicThreadCommentEditAction.Matcher())
	fr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", RequireThreadAndTopic(comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(topicThreadCommentEditActionCancel))))).Methods("POST").MatcherFunc(topicThreadCommentEditActionCancel.Matcher())

	fr.HandleFunc("/preview", handlers.PreviewPage).Methods("POST")

	api := r.PathPrefix("/api/forum").Subrouter()
	api.HandleFunc("/quote/{commentid}", QuoteApi).Methods("GET")
	api.HandleFunc("/thread/{thread}/drafts", handlers.TaskHandler(draftsTask)).Methods("GET", "POST", "DELETE").MatcherFunc(draftsTask.Matcher())
	api.HandleFunc("/share", share.ShareLink).Methods("GET")
}

// Register registers the forum router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("forum", nil, RegisterRoutes)
}
