package user

import (
	"github.com/gorilla/mux"
	"net/http"

	auth "github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/pkg/handlers"
)

// RegisterRoutes attaches user account endpoints to the router.
func RegisterRoutes(r *mux.Router) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.HandleFunc("", userPage).Methods(http.MethodGet)
	ur.HandleFunc("/logout", userLogoutPage).Methods(http.MethodGet)
	ur.HandleFunc("/lang", userLangPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/lang", userLangSaveLanguagesActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSaveLanguages))
	ur.HandleFunc("/lang", userLangSaveLanguagePreferenceActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSaveLanguage))
	ur.HandleFunc("/lang", userLangSaveAllActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/email", userEmailSaveActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailTestActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskTestMail))
	ur.HandleFunc("/paging", userPagingPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/paging", userPagingSaveActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/page-size", userPageSizePage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/page-size", userPageSizeSaveActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/notifications", userNotificationsPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())
	ur.HandleFunc("/notifications/dismiss", userNotificationsDismissActionPage).Methods(http.MethodPost).MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskDismiss))
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods(http.MethodGet).MatcherFunc(auth.RequiresAnAccount())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}
