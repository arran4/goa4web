package admin

import (
	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
	faq "github.com/arran4/goa4web/handlers/faq"
	forum "github.com/arran4/goa4web/handlers/forum"
	languages "github.com/arran4/goa4web/handlers/languages"
	linker "github.com/arran4/goa4web/handlers/linker"
	news "github.com/arran4/goa4web/handlers/news"
	search "github.com/arran4/goa4web/handlers/search"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	writings "github.com/arran4/goa4web/handlers/writings"
	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the admin endpoints to ar. The router is expected to
// already have any required authentication middleware applied.
func RegisterRoutes(ar *mux.Router) {
	nav.RegisterAdminControlCenter("Categories", "/admin/categories", 20)
	nav.RegisterAdminControlCenter("Notifications", "/admin/notifications", 90)
	nav.RegisterAdminControlCenter("Queued Emails", "/admin/email/queue", 110)
	nav.RegisterAdminControlCenter("Email Template", "/admin/email/template", 120)
	nav.RegisterAdminControlCenter("Dead Letter Queue", "/admin/dlq", 130)
	nav.RegisterAdminControlCenter("Server Stats", "/admin/stats", 140)
	nav.RegisterAdminControlCenter("Site Settings", "/admin/settings", 150)
	nav.RegisterAdminControlCenter("Usage Stats", "/admin/usage", 160)

	ar.HandleFunc("", AdminPage).Methods("GET")
	ar.HandleFunc("/", AdminPage).Methods("GET")
	ar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/email/queue", AdminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/queue", AdminEmailQueueResendActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskResend))
	ar.HandleFunc("/email/queue", AdminEmailQueueDeleteActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDelete))
	ar.HandleFunc("/email/template", AdminEmailTemplatePage).Methods("GET")
	ar.HandleFunc("/email/template", AdminEmailTemplateSaveActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUpdate))
	ar.HandleFunc("/email/template", AdminEmailTemplateTestActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskTestMail))
	ar.HandleFunc("/dlq", AdminDLQPage).Methods("GET")
	ar.HandleFunc("/dlq", AdminDLQAction).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDelete))
	ar.HandleFunc("/notifications", AdminNotificationsPage).Methods("GET")
	ar.HandleFunc("/notifications", AdminNotificationsMarkReadActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDismiss))
	ar.HandleFunc("/notifications", AdminNotificationsPurgeActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskPurge))
	ar.HandleFunc("/notifications", AdminNotificationsSendActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskNotify))
	ar.HandleFunc("/user", adminUserListPage).Methods("GET")
	ar.HandleFunc("/user/{id}", adminUserProfilePage).Methods("GET")
	ar.HandleFunc("/announcements", AdminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", AdminAnnouncementsAddActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskAdd))
	ar.HandleFunc("/announcements", AdminAnnouncementsDeleteActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDelete))
	ar.HandleFunc("/ipbans", AdminIPBanPage).Methods("GET")
	ar.HandleFunc("/ipbans", AdminIPBanAddActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskAdd))
	ar.HandleFunc("/ipbans", AdminIPBanDeleteActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDelete))
	ar.HandleFunc("/audit", AdminAuditLogPage).Methods("GET")
	ar.HandleFunc("/settings", AdminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/stats", AdminServerStatsPage).Methods("GET")
	ar.HandleFunc("/usage", AdminUsageStatsPage).Methods("GET")

	// forum admin routes
	forum.RegisterAdminRoutes(ar)

	// linker admin
	linker.RegisterAdminRoutes(ar)

	// faq admin
	faq.RegisterAdminRoutes(ar)
	search.RegisterAdminRoutes(ar)
	userhandlers.RegisterAdminRoutes(ar)
	languages.RegisterAdminRoutes(ar)

	// news admin
	nar := ar.PathPrefix("/news").Subrouter()
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsPage).Methods("GET")
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskAllow))
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskRemoveLower))

	// writings admin
	writings.RegisterAdminRoutes(ar)

	ar.HandleFunc("/reload", AdminReloadConfigPage).Methods("POST")
	ar.HandleFunc("/shutdown", AdminShutdownPage).Methods("POST")
}

// Register registers the admin router module.
func Register() {
	router.RegisterModule("admin", []string{"faq", "forum", "languages", "linker", "news", "search", "user", "writings"}, func(r *mux.Router) {
		ar := r.PathPrefix("/admin").Subrouter()
		ar.Use(router.AdminCheckerMiddleware)
		RegisterRoutes(ar)
	})
}
