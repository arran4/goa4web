package user

import (
	"github.com/gorilla/mux"
	"net/http"

	auth "github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches user account endpoints to the router.
func RegisterRoutes(r *mux.Router) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.HandleFunc("", userPage).Methods(http.MethodGet)
	ur.HandleFunc("/logout", userLogoutPage).Methods(http.MethodGet)
	ur.HandleFunc("/lang", userLangPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/lang", userLangSaveLanguagesActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveLanguagesTask.Match)
	ur.HandleFunc("/lang", userLangSaveLanguagePreferenceActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveLanguageTask.Match)
	ur.HandleFunc("/lang", userLangSaveAllActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllTask.Match)
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/email", userEmailSaveActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllTask.Match)
	ur.HandleFunc("/email/add", userEmailAddActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(AddEmailTask.Match)
	ur.HandleFunc("/email/resend", userEmailResendActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(AddEmailTask.Match)
	ur.HandleFunc("/email/delete", userEmailDeleteActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(DeleteEmailTask.Match)
	ur.HandleFunc("/email/notify", userEmailNotifyActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(AddEmailTask.Match)
	ur.HandleFunc("/email/verify", userEmailVerifyCodePage).Methods(http.MethodGet)
	ur.HandleFunc("/email", userEmailTestActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(TestMailTask.Match)
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/paging", userPagingSaveActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllTask.Match)
	ur.HandleFunc("/page-size", userPageSizePage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/page-size", userPageSizeSaveActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllTask.Match)
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/notifications", userNotificationEmailActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveAllTask.Match)
	ur.HandleFunc("/notifications/dismiss", userNotificationsDismissActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(DismissTask.Match)
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/notifications/gallery", userGalleryPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/subscriptions", userSubscriptionsPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/subscriptions/update", userSubscriptionsUpdateAction).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(UpdateSubscriptionsTask.Match)
	ur.HandleFunc("/subscriptions/delete", userSubscriptionsDeleteAction).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(DeleteTask.Match)

	// legacy redirects
	r.HandleFunc("/user/lang", common.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", common.RedirectPermanent("/usr/email"))
}

// Register registers the user router module.
func Register() {
	router.RegisterModule("user", nil, RegisterRoutes)
}
