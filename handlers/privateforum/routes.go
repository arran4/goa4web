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
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLinkWithViewPermission("Private", "/private", SectionWeight, "privateforum", "topic")
	pr := r.PathPrefix("/private").Subrouter()
	pr.Use(handlers.IndexMiddleware(CustomIndex), handlers.SectionMiddleware("privateforum"), forumhandlers.BasePathMiddleware("/private"))
	pr.HandleFunc("", Page).Methods(http.MethodGet)
	pr.HandleFunc("", handlers.TaskHandler(privateTopicCreateTask)).Methods(http.MethodPost).MatcherFunc(privateTopicCreateTask.Matcher())
	pr.HandleFunc("/private_forum.js", handlers.PrivateForumJS).Methods(http.MethodGet)
	pr.HandleFunc("/topic_labels.js", handlers.TopicLabelsJS).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}", TopicPage).Methods(http.MethodGet)
	pr.HandleFunc("/api/user-exists", UserExistsAPI).Methods(http.MethodPost)

	pr.HandleFunc("/topic/{topic}/subscribe", handlers.TaskHandler(forumhandlers.SubscribeTopicTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.SubscribeTopicTaskHandler.Matcher())
	pr.HandleFunc("/topic/{topic}/unsubscribe", handlers.TaskHandler(forumhandlers.UnsubscribeTopicTaskHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.UnsubscribeTopicTaskHandler.Matcher())

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
	pr.HandleFunc("/topic/{topic}/thread/cancel", forumhandlers.ThreadNewCancelPage).Methods(http.MethodGet)
	pr.HandleFunc("/topic/{topic}/thread/cancel", handlers.TaskHandler(forumhandlers.ThreadNewCancelHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.ThreadNewCancelHandler.Matcher())
	pr.HandleFunc("/topic/{topic}/thread", handlers.TaskHandler(forumhandlers.ThreadNewCancelHandler)).Methods(http.MethodPost).MatcherFunc(forumhandlers.ThreadNewCancelHandler.Matcher())

	pr.Handle("/topic/{topic}/thread/{thread}", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(ThreadPage))).Methods(http.MethodGet)
	pr.Handle("/topic/{topic}/thread/{thread}", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskDoneAutoRefreshPage))).Methods(http.MethodPost)
	pr.Handle("/topic/{topic}/thread/{thread}/reply", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskHandler(forumhandlers.ReplyTaskHandler)))).Methods(http.MethodPost).MatcherFunc(forumhandlers.ReplyTaskHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/reply", forumhandlers.RequireThreadAndTopic(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadReplyCancelHandler)))).Methods(http.MethodPost).MatcherFunc(forumhandlers.TopicThreadReplyCancelHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", forumhandlers.RequireThreadAndTopic(forumcomments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadCommentEditActionHandler))))).Methods(http.MethodPost).MatcherFunc(forumhandlers.TopicThreadCommentEditActionHandler.Matcher())
	pr.Handle("/topic/{topic}/thread/{thread}/comment/{comment}", forumhandlers.RequireThreadAndTopic(forumcomments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(forumhandlers.TopicThreadCommentEditActionCancelHandler))))).Methods(http.MethodPost).MatcherFunc(forumhandlers.TopicThreadCommentEditActionCancelHandler.Matcher())
}

// Register registers the private forum router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("privateforum", []string{"private"}, RegisterRoutes)
}
