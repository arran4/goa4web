package user

import (
	"github.com/arran4/goa4web/internal/tasks"
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
	ur.HandleFunc("/lang", tasks.Action(saveLanguagesTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveLanguagesTask.Matcher())
	ur.HandleFunc("/lang", tasks.Action(saveLanguageTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveLanguageTask.Matcher())
	ur.HandleFunc("/lang", tasks.Action(saveAllTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/email", tasks.Action(saveEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveEmailTask.Matcher())
	ur.HandleFunc("/email/add", tasks.Action(addEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher())
	ur.HandleFunc("/email/resend", tasks.Action(resendEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(resendEmailTask.Matcher())
	ur.HandleFunc("/email/delete", tasks.Action(deleteEmailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteEmailTask.Matcher())
	ur.HandleFunc("/email/notify", addEmailTask.Notify).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(addEmailTask.Matcher())
	ur.HandleFunc("/email/verify", userEmailVerifyCodePage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/email", tasks.Action(testMailTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(testMailTask.Matcher())
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/paging", tasks.Action(pagingSaveTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(pagingSaveTask.Matcher())
	ur.HandleFunc("/page-size", userPageSizePage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/page-size", tasks.Action(pageSizeSaveTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(pageSizeSaveTask.Matcher())
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications", tasks.Action(saveAllTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveAllTask.Matcher())
	ur.HandleFunc("/notifications/dismiss", tasks.Action(dismissTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(dismissTask.Matcher())
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/notifications/gallery", userGalleryPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions", userSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(handlers.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/update", tasks.Action(updateSubscriptionsTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(updateSubscriptionsTask.Matcher())
	ur.HandleFunc("/subscriptions/delete", tasks.Action(deleteTask)).Methods(http.MethodPost).MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(deleteTask.Matcher())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}

// Register registers the user router module.
func Register() {
	router.RegisterModule("user", nil, RegisterRoutes)
}
