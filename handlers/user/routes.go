package user

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches user account endpoints to the router.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, _ *nav.Registry) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	ur.Use(handlers.IndexMiddleware(CustomIndex))
	ur.HandleFunc("", UserPage).Methods(http.MethodGet)
	ur.HandleFunc("/logout", userLogoutPage).Methods(http.MethodGet)
	ur.HandleFunc("/lang", userLangPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/lang", handlers.TaskHandler(saveLanguagesTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveLanguagesTask.Matcher())
	ur.HandleFunc("/lang", handlers.TaskHandler(saveLanguageTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveLanguageTask.Matcher())
	ur.HandleFunc("/lang", handlers.TaskHandler(saveAllTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/email", handlers.TaskHandler(saveEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveEmailTask.Matcher())
	ur.HandleFunc("/email/add", handlers.TaskHandler(addEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher())
	ur.HandleFunc("/email/resend", handlers.TaskHandler(resendVerificationEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(resendVerificationEmailTask.Matcher())
	ur.HandleFunc("/email/delete", handlers.TaskHandler(deleteEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteEmailTask.Matcher())
	ur.HandleFunc("/email/notify", addEmailTask.Notify).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher())
	ur.HandleFunc("/email/verify", userEmailVerifyCodePage).Methods(http.MethodGet, http.MethodPost)
	ur.HandleFunc("/email", handlers.TaskHandler(testMailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(testMailTask.Matcher())
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/paging", handlers.TaskHandler(pagingSaveTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(pagingSaveTask.Matcher())
	ur.HandleFunc("/timezone", userTimezonePage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/timezone", handlers.TaskHandler(saveTimezoneTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveTimezoneTask.Matcher())
	ur.HandleFunc("/appearance", userAppearancePage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/appearance", handlers.TaskHandler(appearanceSaveTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(appearanceSaveTask.Matcher())
	ur.HandleFunc("/profile", userPublicProfileSettingPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/profile", handlers.TaskHandler(publicProfileSaveTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(publicProfileSaveTask.Matcher())
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications", handlers.TaskHandler(saveAllTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/notifications", handlers.TaskHandler(saveDigestTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveDigestTask.Matcher())
	ur.HandleFunc("/notifications", handlers.TaskHandler(sendDigestNowTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(sendDigestNowTask.Matcher())
	ur.HandleFunc("/notifications", handlers.TaskHandler(sendDigestPreviewTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(sendDigestPreviewTask.Matcher())
	ur.HandleFunc("/notifications/dismiss", handlers.TaskHandler(dismissTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(dismissTask.Matcher())
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications/atom", notificationsAtomPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications/open/{id}", userNotificationOpenPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications/go/{id}", notificationsGoPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/u/{username}/notifications/rss", notificationsRssPage).Methods(http.MethodGet)
	ur.HandleFunc("/u/{username}/notifications/atom", notificationsAtomPage).Methods(http.MethodGet)
	ur.HandleFunc("/notifications/gallery", userGalleryPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions", userSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/add", userSubscriptionAddPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/threads", userThreadSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/update", handlers.TaskHandler(updateSubscriptionsTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(updateSubscriptionsTask.Matcher())
	ur.HandleFunc("/subscriptions/delete", handlers.TaskHandler(deleteTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteTask.Matcher())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))

	r.HandleFunc("/user/profile/{username}", userPublicProfilePage).Methods(http.MethodGet)
	r.HandleFunc("/user/profile/{username}/", userPublicProfilePage).Methods(http.MethodGet)

	r.HandleFunc("/user/{user:[0-9]+}/reset", UserResetPasswordPage).Methods("GET")
	r.HandleFunc("/user/{user:[0-9]+}/reset", handlers.TaskHandler(userResetPasswordTask)).Methods("POST").MatcherFunc(userResetPasswordTask.Matcher())
}

// Register registers the user router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("user", nil, RegisterRoutes)
}
