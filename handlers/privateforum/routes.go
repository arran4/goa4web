package privateforum

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	forumcomments "github.com/arran4/goa4web/handlers/forum/comments"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the private forum endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLinkWithViewPermission("Private", "/private", SectionWeight, "privateforum", "topic")
	pr := r.PathPrefix("/private").Subrouter()
	pr.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	pr.Use(handlers.IndexMiddleware(CustomIndex), handlers.SectionMiddleware("privateforum"), forumhandlers.BasePathMiddleware("/private"))
	pr.HandleFunc("", PrivateForumPage).Methods(http.MethodGet)
	pr.HandleFunc("/preview", handlers.PreviewPage).Methods("POST")
	// Dedicated page to start a private group discussion
	pr.HandleFunc("/topic/new", StartGroupDiscussionPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/topic/new", handlers.TaskHandler(privateTopicCreateTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(privateTopicCreateTask.Matcher())
	pr.HandleFunc("", handlers.TaskHandler(privateTopicCreateTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(privateTopicCreateTask.Matcher())
	pr.HandleFunc("/private_forum.js", handlers.PrivateForumJS(cfg)).Methods(http.MethodGet)
	pr.HandleFunc("/topic_labels.js", handlers.TopicLabelsJS(cfg)).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}", TopicPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())

	// Provide GET confirmation pages for subscribe/unsubscribe (mirrors public forum)
	pr.HandleFunc("/topic/{topic}/subscribe", forumhandlers.SubscribeTopicPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/topic/{topic}/unsubscribe", forumhandlers.UnsubscribeTopicPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/topic/{topic}/subscribe", handlers.TaskHandler(forumhandlers.SubscribeTopicTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.SubscribeTopicTaskHandler.Matcher())
	pr.HandleFunc("/topic/{topic}/unsubscribe", handlers.TaskHandler(forumhandlers.UnsubscribeTopicTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.UnsubscribeTopicTaskHandler.Matcher())

	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.MarkThreadReadTaskHandler)).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.SetLabelsTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.SetLabelsTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.AddPublicLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.AddPublicLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.RemovePublicLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.RemovePublicLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.AddAuthorLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.AddAuthorLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.RemoveAuthorLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.RemoveAuthorLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.AddPrivateLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.AddPrivateLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.RemovePrivateLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.RemovePrivateLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.MarkThreadReadTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.MarkThreadReadTaskHandler.Matcher())

	pr.HandleFunc("/topic/{topic}/thread", forumhandlers.CreateThreadTaskHandler.Page).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(forumhandlers.CreateThreadTaskHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.CreateThreadTaskHandler.Matcher())
	// Backwards-compatible alias: `/private/topic/{topic}/cancel` → `/private/topic/{topic}/thread/cancel`
	pr.HandleFunc("/topic/{topic}/cancel", TopicCancelAlias).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/topic/{topic}/thread/cancel", forumhandlers.ThreadNewCancelPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.HandleFunc("/topic/{topic}/thread/cancel", handlers.TaskHandler(forumhandlers.ThreadNewCancelHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.ThreadNewCancelHandler.Matcher())
	pr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(forumhandlers.ThreadNewCancelHandler)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.ThreadNewCancelHandler.Matcher())

	// OpenGraph preview endpoints (no auth required for social media bots if signed)
	pr.HandleFunc("/shared/topic/{topic}", SharedTopicPreviewPage).Methods(http.MethodGet, http.MethodHead)
	pr.HandleFunc("/shared/topic/{topic}/ts/{ts}/sign/{sign}", SharedTopicPreviewPage).Methods(http.MethodGet, http.MethodHead)
	pr.HandleFunc("/shared/topic/{topic}/thread/{thread}", SharedThreadPreviewPage).Methods(http.MethodGet, http.MethodHead)
	pr.HandleFunc("/shared/topic/{topic}/thread/{thread}/ts/{ts}/sign/{sign}", SharedThreadPreviewPage).Methods(http.MethodGet, http.MethodHead)

	pr.Handle("/topic/{topic}/thread/{thread}", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	pr.Handle("/topic/{topic}/thread/{thread}", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount())
	pr.Handle("/topic/{topic}/thread/{thread}/reply", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskHandler(forumhandlers.ReplyTaskHandler)))).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.ReplyTaskHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", forumhandlers.RequireThreadAndTopic(forumcomments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadCommentEditActionHandler))))).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.TopicThreadCommentEditActionHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", forumhandlers.RequireThreadAndTopic(forumcomments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadCommentEditActionCancelHandler))))).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(forumhandlers.TopicThreadCommentEditActionCancelHandler.Matcher())

}

// Register registers the private forum router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("privateforum", nil, RegisterRoutes)
}
