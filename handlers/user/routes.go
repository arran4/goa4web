package user

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/handlers"
	auth "github.com/arran4/goa4web/handlers/auth"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches user account endpoints to the router.
func RegisterRoutes(r *mux.Router) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.HandleFunc("", userPage).Methods(http.MethodGet)
	ur.HandleFunc("/logout", userLogoutPage).Methods(http.MethodGet)
	ur.HandleFunc("/lang", userLangPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/lang", saveLanguagesTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(saveLanguagesTask.Matcher())
	ur.HandleFunc("/lang", saveLanguageTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveLanguageEvent.Match)
	ur.HandleFunc("/lang", saveLangAllTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllEvent.Match)
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/email", emailSaveTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllEvent.Match)
	ur.HandleFunc("/email/add", emailAddTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(AddEmailEvent.Match)
	ur.HandleFunc("/email/resend", emailResendTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(AddEmailEvent.Match)
	ur.HandleFunc("/email/delete", emailDeleteTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(DeleteEmailEvent.Match)
	ur.HandleFunc("/email/notify", emailNotifyTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(AddEmailEvent.Match)
	ur.HandleFunc("/email/verify", userEmailVerifyCodePage).Methods(http.MethodGet)
	ur.HandleFunc("/email", emailTestTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(TestMailEvent.Match)
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/paging", pagingSaveTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllEvent.Match)
	ur.HandleFunc("/page-size", userPageSizePage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/page-size", pageSizeSaveTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllEvent.Match)
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/notifications", notificationEmailTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllEvent.Match)
	ur.HandleFunc("/notifications/dismiss", notificationDismissTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(DismissEvent.Match)
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/notifications/gallery", userGalleryPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/subscriptions", userSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/update", subscriptionsUpdateTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(UpdateSubscriptionsEvent.Match)
	ur.HandleFunc("/subscriptions/delete", subscriptionsDeleteTask.Action).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(DeleteEvent.Match)

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}

// Register registers the user router module.
func Register() {
	router.RegisterModule("user", nil, RegisterRoutes)
}
