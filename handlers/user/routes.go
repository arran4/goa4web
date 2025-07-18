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
	ur.HandleFunc("", userPage).Methods(http.MethodGet)
	ur.HandleFunc("/logout", userLogoutPage).Methods(http.MethodGet)
	ur.HandleFunc("/lang", userLangPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/lang", saveLanguagesTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveLanguagesTask.Matcher())
	ur.HandleFunc("/lang", saveLanguageTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveLanguageTask.Matcher())
	ur.HandleFunc("/lang", saveAllTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/email", saveEmailTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveEmailTask.Matcher())
	ur.HandleFunc("/email/add", addEmailTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher())
	ur.HandleFunc("/email/resend", addEmailTask.Resend).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher()) // TODO resend should be it's own task.
	ur.HandleFunc("/email/delete", deleteEmailTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteEmailTask.Matcher())
	ur.HandleFunc("/email/notify", addEmailTask.Notify).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher())
	ur.HandleFunc("/email/verify", userEmailVerifyCodePage).Methods(http.MethodGet)
	ur.HandleFunc("/email", testMailTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(testMailTask.Matcher())
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/paging", pagingSaveTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(pagingSaveTask.Matcher())
	ur.HandleFunc("/page-size", userPageSizePage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/page-size", pageSizeSaveTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(pageSizeSaveTask.Matcher())
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications", saveAllTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/notifications/dismiss", dismissTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(dismissTask.Matcher())
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications/gallery", userGalleryPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions", userSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/update", updateSubscriptionsTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(updateSubscriptionsTask.Matcher())
	ur.HandleFunc("/subscriptions/delete", deleteTask.Action).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteTask.Matcher())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}

// Register registers the user router module.
func Register() {
	router.RegisterModule("user", nil, RegisterRoutes)
}
