package admin

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	blogs "github.com/arran4/goa4web/handlers/blogs"
	faq "github.com/arran4/goa4web/handlers/faq"
	forum "github.com/arran4/goa4web/handlers/forum"
	imagebbs "github.com/arran4/goa4web/handlers/imagebbs"
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
func RegisterRoutes(ar *mux.Router, navReg *nav.Registry) {
	navReg.RegisterAdminControlCenter("Categories", "/admin/categories", 20)
	navReg.RegisterAdminControlCenter("Notifications", "/admin/notifications", 90)
	navReg.RegisterAdminControlCenter("Queued Emails", "/admin/email/queue", 110)
	navReg.RegisterAdminControlCenter("Sent Emails", "/admin/email/sent", 115)
	navReg.RegisterAdminControlCenter("Email Template", "/admin/email/template", 120)
	navReg.RegisterAdminControlCenter("Dead Letter Queue", "/admin/dlq", 130)
	navReg.RegisterAdminControlCenter("Server Stats", "/admin/stats", 140)
	navReg.RegisterAdminControlCenter("Requests", "/admin/requests", 145)
	navReg.RegisterAdminControlCenter("Site Settings", "/admin/settings", 150)
	navReg.RegisterAdminControlCenter("Pagination", "/admin/page-size", 152)
	navReg.RegisterAdminControlCenter("Usage Stats", "/admin/usage", 160)

	ar.HandleFunc("", AdminPage).Methods("GET")
	ar.HandleFunc("/", AdminPage).Methods("GET")
	ar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/email/queue", AdminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/sent", AdminSentEmailsPage).Methods("GET")
	ar.HandleFunc("/email/queue", handlers.TaskHandler(resendQueueTask)).Methods("POST").MatcherFunc(resendQueueTask.Matcher())
	ar.HandleFunc("/email/queue", handlers.TaskHandler(deleteQueueTask)).Methods("POST").MatcherFunc(deleteQueueTask.Matcher())
	ar.HandleFunc("/email/template", AdminEmailTemplatePage).Methods("GET")
	ar.HandleFunc("/email/template", handlers.TaskHandler(saveTemplateTask)).Methods("POST").MatcherFunc(saveTemplateTask.Matcher())
	ar.HandleFunc("/email/template", handlers.TaskHandler(testTemplateTask)).Methods("POST").MatcherFunc(testTemplateTask.Matcher())
	ar.HandleFunc("/dlq", AdminDLQPage).Methods("GET")
	ar.HandleFunc("/dlq", handlers.TaskHandler(deleteDLQTask)).Methods("POST").MatcherFunc(deleteDLQTask.Matcher())
	ar.HandleFunc("/notifications", AdminNotificationsPage).Methods("GET")
	ar.HandleFunc("/notifications", handlers.TaskHandler(markReadTask)).Methods("POST").MatcherFunc(markReadTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(markUnreadTask)).Methods("POST").MatcherFunc(markUnreadTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(toggleNotificationReadTask)).Methods("POST").MatcherFunc(toggleNotificationReadTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(purgeSelectedNotificationsTask)).Methods("POST").MatcherFunc(purgeSelectedNotificationsTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(purgeReadNotificationsTask)).Methods("POST").MatcherFunc(purgeReadNotificationsTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(sendNotificationTask)).Methods("POST").MatcherFunc(sendNotificationTask.Matcher())
	ar.HandleFunc("/requests", AdminRequestQueuePage).Methods("GET")
	ar.HandleFunc("/requests/archive", AdminRequestArchivePage).Methods("GET")
	ar.HandleFunc("/request/{id}", adminRequestPage).Methods("GET")
	ar.HandleFunc("/request/{id}/comment", adminRequestAddCommentPage).Methods("POST")
	ar.HandleFunc("/request/{id}/accept", handlers.TaskHandler(acceptRequestTask)).Methods("POST").MatcherFunc(acceptRequestTask.Matcher())
	ar.HandleFunc("/request/{id}/reject", handlers.TaskHandler(rejectRequestTask)).Methods("POST").MatcherFunc(rejectRequestTask.Matcher())
	ar.HandleFunc("/request/{id}/query", handlers.TaskHandler(queryRequestTask)).Methods("POST").MatcherFunc(queryRequestTask.Matcher())
	ar.HandleFunc("/user", adminUserListPage).Methods("GET")
	ar.HandleFunc("/user/{id}", adminUserProfilePage).Methods("GET")
	ar.HandleFunc("/user/{id}/comment", adminUserAddCommentPage).Methods("POST")
	ar.HandleFunc("/announcements", AdminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", handlers.TaskHandler(addAnnouncementTask)).Methods("POST").MatcherFunc(addAnnouncementTask.Matcher())
	ar.HandleFunc("/announcements", handlers.TaskHandler(deleteAnnouncementTask)).Methods("POST").MatcherFunc(deleteAnnouncementTask.Matcher())
	ar.HandleFunc("/ipbans", AdminIPBanPage).Methods("GET")
	ar.HandleFunc("/ipbans", handlers.TaskHandler(addIPBanTask)).Methods("POST").MatcherFunc(addIPBanTask.Matcher())
	ar.HandleFunc("/ipbans", handlers.TaskHandler(deleteIPBanTask)).Methods("POST").MatcherFunc(deleteIPBanTask.Matcher())
	ar.HandleFunc("/audit", AdminAuditLogPage).Methods("GET")
	ar.HandleFunc("/settings", AdminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/page-size", AdminPageSizePage).Methods("GET", "POST")
	ar.HandleFunc("/stats", AdminServerStatsPage).Methods("GET")
	ar.HandleFunc("/usage", AdminUsageStatsPage).Methods("GET")

	// forum admin routes
	forum.RegisterAdminRoutes(ar)

	// imagebbs admin
	imagebbs.RegisterAdminRoutes(ar)

	// linker admin
	linker.RegisterAdminRoutes(ar)

	// faq admin
	faq.RegisterAdminRoutes(ar)
	search.RegisterAdminRoutes(ar)
	userhandlers.RegisterAdminRoutes(ar, navReg)
	languages.RegisterAdminRoutes(ar, navReg)
	blogs.RegisterAdminRoutes(ar)

	// news admin
	nar := ar.PathPrefix("/news").Subrouter()
	nar.HandleFunc("/users/roles", news.AdminUserRolesPage).Methods("GET")
	nar.HandleFunc("/users/roles", handlers.TaskHandler(newsUserAllow)).Methods("POST").MatcherFunc(newsUserAllow.Matcher())
	nar.HandleFunc("/users/roles", handlers.TaskHandler(newsUserRemove)).Methods("POST").MatcherFunc(newsUserRemove.Matcher())

	// writings admin
	writings.RegisterAdminRoutes(ar)

	// TODO ensure requires administrator to run. Even from CLI.
	ar.HandleFunc("/reload", AdminReloadConfigPage).
		Methods("POST").
		MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/shutdown", handlers.TaskHandler(serverShutdownTask)).
		Methods("POST").
		MatcherFunc(handlers.RequiredAccess("administrator")).
		MatcherFunc(serverShutdownTask.Matcher())

	api := ar.PathPrefix("/api").Subrouter()
	api.Use(router.AdminCheckerMiddleware)
	api.HandleFunc("/shutdown", AdminAPIServerShutdown).MatcherFunc(AdminAPISigned()).Methods("POST")
}

// Register registers the admin router module.
func Register(reg *router.Registry, navReg *nav.Registry) {
	reg.RegisterModule("admin", []string{"faq", "forum", "imagebbs", "languages", "linker", "news", "search", "user", "writings", "blogs"}, func(r *mux.Router) {
		ar := r.PathPrefix("/admin").Subrouter()
		ar.Use(router.AdminCheckerMiddleware)
		ar.Use(handlers.IndexMiddleware(CustomIndex))
		RegisterRoutes(ar, navReg)
	})
}
