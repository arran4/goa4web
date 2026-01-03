package privateforum

import (
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/arran4/gorillamuxlogic"

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
	pr.Use(handlers.IndexMiddleware(CustomIndex), handlers.SectionMiddleware("privateforum"), forumhandlers.BasePathMiddleware("/private"))
	pr.HandleFunc("", PrivateForumPage).Methods(http.MethodGet)
	pr.HandleFunc("/preview", handlers.PreviewPage).Methods("POST")
	// Dedicated page to start a private group discussion
	pr.HandleFunc("/topic/new", StartGroupDiscussionPage).Methods(http.MethodGet)
	pr.HandleFunc("/topic/new", handlers.TaskHandler(privateTopicCreateTask)).Methods(http.MethodPost).MatcherFunc(privateTopicCreateTask.Matcher())
	pr.HandleFunc("", handlers.TaskHandler(privateTopicCreateTask)).Methods(http.MethodPost).MatcherFunc(privateTopicCreateTask.Matcher())
	pr.HandleFunc("/private_forum.js", handlers.PrivateForumJS(cfg)).Methods(http.MethodGet)
	pr.HandleFunc("/topic_labels.js", handlers.TopicLabelsJS(cfg)).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}", TopicPage).Methods(http.MethodGet)

	// Provide GET confirmation pages for subscribe/unsubscribe (mirrors public forum)
	pr.HandleFunc("/topic/{topic}/subscribe", forumhandlers.SubscribeTopicPage).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}/unsubscribe", forumhandlers.UnsubscribeTopicPage).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}/subscribe", handlers.TaskHandler(forumhandlers.SubscribeTopicTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.SubscribeTopicTaskHandler.Matcher())
	pr.HandleFunc("/topic/{topic}/unsubscribe", handlers.TaskHandler(forumhandlers.UnsubscribeTopicTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.UnsubscribeTopicTaskHandler.Matcher())

	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.MarkThreadReadTaskHandler)).Methods(http.MethodGet)
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.SetLabelsTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.SetLabelsTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.AddPublicLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.AddPublicLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.RemovePublicLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.RemovePublicLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.AddAuthorLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.AddAuthorLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.RemoveAuthorLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.RemoveAuthorLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.AddPrivateLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.AddPrivateLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.RemovePrivateLabelTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.RemovePrivateLabelTaskHandler.Matcher())
	pr.HandleFunc("/thread/{thread}/labels", handlers.TaskHandler(forumhandlers.MarkThreadReadTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.MarkThreadReadTaskHandler.Matcher())

	pr.HandleFunc("/topic/{topic}/thread", forumhandlers.CreateThreadTaskHandler.Page).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(forumhandlers.CreateThreadTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.CreateThreadTaskHandler.Matcher())
	// Backwards-compatible alias: `/private/topic/{topic}/cancel` → `/private/topic/{topic}/thread/cancel`
	pr.HandleFunc("/topic/{topic}/cancel", TopicCancelAlias).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}/thread/cancel", forumhandlers.ThreadNewCancelPage).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}/thread/cancel", handlers.TaskHandler(forumhandlers.ThreadNewCancelHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.ThreadNewCancelHandler.Matcher())
	pr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(forumhandlers.ThreadNewCancelHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.ThreadNewCancelHandler.Matcher())

	pr.Handle("/topic/{topic}/thread/{thread}", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods(http.MethodGet)
	pr.Handle("/topic/{topic}/thread/{thread}", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods(http.MethodPost)
	pr.Handle("/topic/{topic}/thread/{thread}/reply", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskHandler(forumhandlers.ReplyTaskHandler)))).Methods(http.MethodPost).MatcherFunc(forumhandlers.ReplyTaskHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", forumhandlers.RequireThreadAndTopic(forumcomments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadCommentEditActionHandler))))).Methods(http.MethodPost).MatcherFunc(forumhandlers.TopicThreadCommentEditActionHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", forumhandlers.RequireThreadAndTopic(forumcomments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadCommentEditActionCancelHandler))))).Methods(http.MethodPost).MatcherFunc(forumhandlers.TopicThreadCommentEditActionCancelHandler.Matcher())

	pr.HandleFunc("/{path:.*}", handlers.RenderPermissionDenied).MatcherFunc(Not(handlers.RequiresAnAccount()))
}

// Register registers the private forum router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("privateforum", nil, RegisterRoutes)
}
