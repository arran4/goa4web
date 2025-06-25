package user

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/pkg/handlers"
)

// RegisterRoutes attaches user account endpoints to the router.
func RegisterRoutes(r *mux.Router) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.HandleFunc("", userPage).Methods(http.MethodGet)
	ur.HandleFunc("/logout", userLogoutPage).Methods(http.MethodGet)
	ur.HandleFunc("/lang", userLangPage).Methods(http.MethodGet).MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/lang", userLangSaveLanguagesActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveLanguages))
	ur.HandleFunc("/lang", userLangSaveLanguagePreferenceActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveLanguage))
	ur.HandleFunc("/lang", userLangSaveAllActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/email", userEmailSaveActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailTestActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskTestMail))
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/paging", userPagingSaveActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/page-size", userPageSizePage).Methods(http.MethodGet).MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/page-size", userPageSizeSaveActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/notifications/dismiss", userNotificationsDismissActionPage).Methods(http.MethodPost).MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskDismiss))
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(RequiresAnAccount())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}
