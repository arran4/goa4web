package user

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches user account endpoints to the router.
func RegisterRoutes(r *mux.Router) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.Use(handlers.IndexMiddleware(CustomIndex))
	ur.HandleFunc("", userPage).Methods(http.MethodGet)
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
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications", handlers.TaskHandler(saveAllTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/notifications/dismiss", handlers.TaskHandler(dismissTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(dismissTask.Matcher())
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications/gallery", userGalleryPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions", userSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/update", handlers.TaskHandler(updateSubscriptionsTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(updateSubscriptionsTask.Matcher())
	ur.HandleFunc("/subscriptions/delete", handlers.TaskHandler(deleteTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteTask.Matcher())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}

// Register registers the user router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("user", nil, RegisterRoutes)
}
