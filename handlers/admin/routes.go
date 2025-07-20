package admin

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"

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
	nav.RegisterAdminControlCenter("Information", "/admin/information", InformationSectionWeight)
	nav.RegisterAdminControlCenter("Site Settings", "/admin/settings", 150)
	nav.RegisterAdminControlCenter("Usage Stats", "/admin/usage", 160)

	ar.HandleFunc("", AdminPage).Methods("GET")
	ar.HandleFunc("/", AdminPage).Methods("GET")
	ar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/email/queue", AdminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/queue", tasks.Action(resendQueueTask)).Methods("POST").MatcherFunc(resendQueueTask.Matcher())
	ar.HandleFunc("/email/queue", tasks.Action(deleteQueueTask)).Methods("POST").MatcherFunc(deleteQueueTask.Matcher())
	ar.HandleFunc("/email/template", AdminEmailTemplatePage).Methods("GET")
	ar.HandleFunc("/email/template", tasks.Action(saveTemplateTask)).Methods("POST").MatcherFunc(saveTemplateTask.Matcher())
	ar.HandleFunc("/email/template", tasks.Action(testTemplateTask)).Methods("POST").MatcherFunc(testTemplateTask.Matcher())
	ar.HandleFunc("/dlq", AdminDLQPage).Methods("GET")
	ar.HandleFunc("/dlq", tasks.Action(deleteDLQTask)).Methods("POST").MatcherFunc(deleteDLQTask.Matcher())
	ar.HandleFunc("/notifications", AdminNotificationsPage).Methods("GET")
	ar.HandleFunc("/notifications", tasks.Action(markReadTask)).Methods("POST").MatcherFunc(markReadTask.Matcher())
	ar.HandleFunc("/notifications", tasks.Action(purgeNotificationsTask)).Methods("POST").MatcherFunc(purgeNotificationsTask.Matcher())
	ar.HandleFunc("/notifications", tasks.Action(sendNotificationTask)).Methods("POST").MatcherFunc(sendNotificationTask.Matcher())
	ar.HandleFunc("/user", adminUserListPage).Methods("GET")
	ar.HandleFunc("/user/{id}", adminUserProfilePage).Methods("GET")
	ar.HandleFunc("/user/{id}/comment", adminUserAddCommentPage).Methods("POST")
	ar.HandleFunc("/announcements", AdminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", tasks.Action(addAnnouncementTask)).Methods("POST").MatcherFunc(addAnnouncementTask.Matcher())
	ar.HandleFunc("/announcements", tasks.Action(deleteAnnouncementTask)).Methods("POST").MatcherFunc(deleteAnnouncementTask.Matcher())
	ar.HandleFunc("/ipbans", AdminIPBanPage).Methods("GET")
	ar.HandleFunc("/ipbans", tasks.Action(addIPBanTask)).Methods("POST").MatcherFunc(addIPBanTask.Matcher())
	ar.HandleFunc("/ipbans", tasks.Action(deleteIPBanTask)).Methods("POST").MatcherFunc(deleteIPBanTask.Matcher())
	ar.HandleFunc("/audit", AdminAuditLogPage).Methods("GET")
	ar.HandleFunc("/settings", AdminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/stats", AdminServerStatsPage).Methods("GET")
	ar.HandleFunc("/information", AdminInformationPage).Methods("GET")
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
	nar.HandleFunc("/users/roles", news.AdminUserRolesPage).Methods("GET")
	nar.HandleFunc("/users/roles", tasks.Action(newsUserAllow)).Methods("POST").MatcherFunc(newsUserAllow.Matcher())
	nar.HandleFunc("/users/roles", tasks.Action(newsUserRemove)).Methods("POST").MatcherFunc(newsUserRemove.Matcher())

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
		ar.Use(handlers.IndexMiddleware(CustomIndex))
		RegisterRoutes(ar)
	})
}
