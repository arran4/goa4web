package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/handlers/faq"
	"github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/handlers/imagebbs"
	"github.com/arran4/goa4web/handlers/languages"
	"github.com/arran4/goa4web/handlers/linker"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/handlers/search"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/handlers/writings"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/tasks"
)

// RegisterRoutes attaches the admin endpoints to ar. The router is expected to
// already have any required authentication middleware applied.
func (h *Handlers) RegisterRoutes(ar *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
	ar.Use(handlers.SectionMiddleware("admin"))
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Roles"), "Roles", "/admin/roles", 25)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Roles"), "Role SQL Loader", "/admin/roles/load", 26)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Roles"), "Role Templates", "/admin/roles/templates", 26)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Grants"), "Grants", "/admin/grants", 27)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Grants"), "Available Grants", "/admin/grants/available", 28)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Maintenance"), "Once Off & Maintenance", "/admin/maintenance", 29)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Maintenance"), "External Links", "/admin/external-links", 30)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "DB"), "DB Status", "/admin/db/status", 31)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "DB"), "DB Schema", "/admin/db/schema", 156)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "DB"), "DB Migrations", "/admin/db/migrations", 157)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Maintenance"), "Link Tools", "/admin/links/tools", 31)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Maintenance"), "Link Remap", "/admin/link-remap", 32)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Notifications"), "Notifications", "/admin/notifications", 90)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Notifications"), "Subscription Templates", "/admin/subscriptions/templates", 95)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Queued Emails", "/admin/email/queue", 110)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Failed Emails", "/admin/email/failed", 112)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Sent Emails", "/admin/email/sent", 115)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Email Tester", "/admin/email/test", 118)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Email Template", "/admin/email/template", 120)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Template Export", "/admin/templates/export", 121)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Email"), "Dead Letter Queue", "/admin/dlq", 130)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Stats"), "Server Stats", "/admin/stats", 140)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Users"), "Requests", "/admin/requests", 145)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Users"), "Password Resets", "/admin/password_resets", 146)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Users"), "Comments", "/admin/comments", 147)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Users"), "Deactivated Comments", "/admin/comments/deactivated", 148)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Settings"), "Site Settings", "/admin/settings", 150)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Settings"), "Config Export", "/admin/config/as-cli", 151)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Settings"), "Config Explain", "/admin/config/explain", 151)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Settings"), "Pagination", "/admin/page-size", 152)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Files"), "Files", "/admin/files", 153)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Files"), "Image Cache", "/admin/images/cache", 154)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "DB"), "DB Backup", "/admin/db/backup", 154)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "DB"), "DB Restore", "/admin/db/restore", 155)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Settings"), "Share Tools", "/admin/share/tools", 155)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Core", "Stats"), "Usage Stats", "/admin/usage", 160)

	ar.HandleFunc("", AdminPage).Methods("GET")
	ar.HandleFunc("/", AdminPage).Methods("GET")
	ar.HandleFunc("/role-grants-editor.js", handlers.RoleGrantsEditorJS(cfg)).Methods(http.MethodGet, http.MethodHead, http.MethodOptions)
	ar.HandleFunc("/grant-add.js", handlers.GrantAddJS(cfg)).Methods(http.MethodGet, http.MethodHead, http.MethodOptions)
	ar.HandleFunc("/roles", AdminRolesPage).Methods("GET")
	ar.HandleFunc("/roles/load", h.AdminRoleLoadPage).Methods("GET", "POST")
	ar.HandleFunc("/roles/templates", AdminRoleTemplatesPage).Methods("GET")
	ar.HandleFunc("/roles/templates", handlers.TaskHandler(roleTemplateApplyTask)).Methods("POST").MatcherFunc(roleTemplateApplyTask.Matcher())
	ar.HandleFunc("/roles", handlers.TaskHandler(rolePublicProfileTask)).Methods("POST").MatcherFunc(rolePublicProfileTask.Matcher())
	ar.HandleFunc("/external-links", AdminExternalLinksPage).Methods("GET")
	ar.HandleFunc("/external-links", handlers.TaskHandler(refreshExternalLinkTask)).Methods("POST").MatcherFunc(refreshExternalLinkTask.Matcher())
	ar.HandleFunc("/external-links", handlers.TaskHandler(deleteExternalLinkTask)).Methods("POST").MatcherFunc(deleteExternalLinkTask.Matcher())
	ar.HandleFunc("/external-links/{id}", AdminExternalLinkDetailsPage).Methods("GET")
	ar.HandleFunc("/external-links/{id}", handlers.TaskHandler(refreshExternalLinkTask)).Methods("POST").MatcherFunc(refreshExternalLinkTask.Matcher())
	ar.HandleFunc("/external-links/{id}", handlers.TaskHandler(deleteExternalLinkDetailTask)).Methods("POST").MatcherFunc(deleteExternalLinkDetailTask.Matcher())
	ar.HandleFunc("/external-links/{id}", handlers.TaskHandler(updateExternalLinkMetadataTask)).Methods("POST").MatcherFunc(updateExternalLinkMetadataTask.Matcher())
	ar.HandleFunc("/db/status", h.AdminDBStatusPage).Methods("GET")
	ar.HandleFunc("/db/schema", h.AdminDBSchemaPage).Methods("GET")
	ar.HandleFunc("/db/migrations", h.AdminDBMigrationsPage).Methods("GET")
	dbSeedTask := h.NewDBSeedTask()
	ar.HandleFunc("/db/status", handlers.TaskHandler(dbSeedTask)).Methods("POST").MatcherFunc(dbSeedTask.Matcher())
	ar.HandleFunc("/links/tools", AdminLinksToolsPage).Methods("GET", "POST")
	ar.HandleFunc("/link-remap", AdminLinkRemapPage).Methods("GET")
	ar.HandleFunc("/link-remap", handlers.TaskHandler(applyLinkRemapTask)).Methods("POST").MatcherFunc(applyLinkRemapTask.Matcher())
	ar.HandleFunc("/db/backup", AdminDBBackupPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/db/backup", handlers.TaskHandler(dbBackupTask)).Methods("POST").MatcherFunc(dbBackupTask.Matcher()).MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/db/restore", AdminDBRestorePage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/db/restore", handlers.TaskHandler(dbRestoreTask)).Methods("POST").MatcherFunc(dbRestoreTask.Matcher()).MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/email/queue", AdminEmailPage).Methods("GET")
	ar.HandleFunc("/email/failed", AdminEmailPage).Methods("GET")
	ar.HandleFunc("/email/sent", AdminEmailPage).Methods("GET")
	ar.HandleFunc("/email/failed", handlers.TaskHandler(resendQueueTask)).Methods("POST").MatcherFunc(resendQueueTask.Matcher())
	ar.HandleFunc("/email/failed", handlers.TaskHandler(deleteQueueTask)).Methods("POST").MatcherFunc(deleteQueueTask.Matcher())
	ar.HandleFunc("/email/sent", handlers.TaskHandler(resendSentEmailTask)).Methods("POST").MatcherFunc(resendSentEmailTask.Matcher())
	ar.HandleFunc("/email/sent", handlers.TaskHandler(retrySentEmailTask)).Methods("POST").MatcherFunc(retrySentEmailTask.Matcher())
	ar.HandleFunc("/email/test", h.AdminEmailTestPage).Methods("GET", "POST")
	ar.HandleFunc("/email/queue", handlers.TaskHandler(resendQueueTask)).Methods("POST").MatcherFunc(resendQueueTask.Matcher())
	ar.HandleFunc("/email/queue", handlers.TaskHandler(deleteQueueTask)).Methods("POST").MatcherFunc(deleteQueueTask.Matcher())
	ar.HandleFunc("/email/template", AdminEmailTemplatePage).Methods("GET")
	ar.HandleFunc("/email/template", handlers.TaskHandler(saveTemplateTask)).Methods("POST").MatcherFunc(saveTemplateTask.Matcher())
	ar.HandleFunc("/email/template", handlers.TaskHandler(testTemplateTask)).Methods("POST").MatcherFunc(testTemplateTask.Matcher())
	ar.HandleFunc("/templates/export", AdminTemplateExportPage).Methods("GET")
	ar.HandleFunc("/templates/export", handlers.TaskHandler(exportTemplatesTask)).Methods("POST").MatcherFunc(exportTemplatesTask.Matcher())
	ar.HandleFunc("/dlq", AdminDLQPage).Methods("GET")
	ar.HandleFunc("/dlq/{provider}/{id}", AdminDLQDetailsPage).Methods("GET")
	ar.HandleFunc("/dlq", handlers.TaskHandler(deleteDLQTask)).Methods("POST").MatcherFunc(deleteDLQTask.Matcher())
	ar.HandleFunc("/dlq", handlers.TaskHandler(reEnlistDLQTask)).Methods("POST").MatcherFunc(reEnlistDLQTask.Matcher())
	ar.HandleFunc("/dlq", handlers.TaskHandler(updateDLQTask)).Methods("POST").MatcherFunc(updateDLQTask.Matcher())
	ar.HandleFunc("/dlq", handlers.TaskHandler(purgeDLQTask)).Methods("POST").MatcherFunc(purgeDLQTask.Matcher())
	ar.HandleFunc("/notifications", AdminNotificationsPage).Methods("GET")
	ar.HandleFunc("/notifications", handlers.TaskHandler(markReadTask)).Methods("POST").MatcherFunc(markReadTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(markUnreadTask)).Methods("POST").MatcherFunc(markUnreadTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(toggleNotificationReadTask)).Methods("POST").MatcherFunc(toggleNotificationReadTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(purgeSelectedNotificationsTask)).Methods("POST").MatcherFunc(purgeSelectedNotificationsTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(purgeReadNotificationsTask)).Methods("POST").MatcherFunc(purgeReadNotificationsTask.Matcher())
	ar.HandleFunc("/notifications", handlers.TaskHandler(sendNotificationTask)).Methods("POST").MatcherFunc(sendNotificationTask.Matcher())
	ar.HandleFunc("/subscriptions/templates", AdminSubscriptionTemplatesPage).Methods("GET")
	ar.HandleFunc("/subscriptions/templates", AdminSubscriptionTemplatesPage).Methods("POST").MatcherFunc(tasks.HasNoTask())
	applySubscriptionTemplateTask := h.NewApplySubscriptionTemplateTask()
	ar.HandleFunc("/subscriptions/templates", handlers.TaskHandler(applySubscriptionTemplateTask)).Methods("POST").MatcherFunc(applySubscriptionTemplateTask.Matcher())
	ar.HandleFunc("/requests", AdminRequestQueuePage).Methods("GET")
	ar.HandleFunc("/requests/archive", AdminRequestArchivePage).Methods("GET")
	ar.HandleFunc("/request/{request}", adminRequestPage).Methods("GET")
	ar.HandleFunc("/request/{request}/comment", adminRequestAddCommentPage).Methods("POST")
	ar.HandleFunc("/request/{request}/accept", handlers.TaskHandler(acceptRequestTask)).Methods("POST").MatcherFunc(acceptRequestTask.Matcher())
	ar.HandleFunc("/request/{request}/reject", handlers.TaskHandler(rejectRequestTask)).Methods("POST").MatcherFunc(rejectRequestTask.Matcher())
	ar.HandleFunc("/request/{request}/dismiss", handlers.TaskHandler(dismissRequestTask)).Methods("POST").MatcherFunc(dismissRequestTask.Matcher())
	ar.HandleFunc("/request/{request}/query", handlers.TaskHandler(queryRequestTask)).Methods("POST").MatcherFunc(queryRequestTask.Matcher())
	ar.HandleFunc("/password_resets", handlers.TaskHandler(clearExpiredPasswordResetsTask)).Methods("POST").MatcherFunc(clearExpiredPasswordResetsTask.Matcher())
	ar.HandleFunc("/password_resets", handlers.TaskHandler(clearUserPasswordResetsTask)).Methods("POST").MatcherFunc(clearUserPasswordResetsTask.Matcher())
	ar.HandleFunc("/password_resets", adminPasswordResetListPage).Methods("GET", "POST")
	ar.HandleFunc("/user", adminUserListPage).Methods("GET")
	ar.HandleFunc("/user/{user}", adminUserProfilePage).Methods("GET")
	ar.HandleFunc("/user/{user}", handlers.TaskHandler(adminAddEmailTask)).Methods("POST").MatcherFunc(adminAddEmailTask.Matcher())
	ar.HandleFunc("/user/{user}", handlers.TaskHandler(adminDeleteEmailTask)).Methods("POST").MatcherFunc(adminDeleteEmailTask.Matcher())
	ar.HandleFunc("/user/{user}", handlers.TaskHandler(adminVerifyEmailTask)).Methods("POST").MatcherFunc(adminVerifyEmailTask.Matcher())
	ar.HandleFunc("/user/{user}", handlers.TaskHandler(adminUnverifyEmailTask)).Methods("POST").MatcherFunc(adminUnverifyEmailTask.Matcher())
	ar.HandleFunc("/user/{user}", handlers.TaskHandler(adminResendVerificationEmailTask)).Methods("POST").MatcherFunc(adminResendVerificationEmailTask.Matcher())
	ar.HandleFunc("/user/{user}/blogs", adminUserBlogsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/writings", adminUserWritingsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/linker", adminUserLinkerPage).Methods("GET")
	ar.HandleFunc("/user/{user}/imagebbs", adminUserImagebbsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/forum", adminUserForumPage).Methods("GET")
	ar.HandleFunc("/user/{user}/comments", adminUserCommentsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/subscriptions", adminUserSubscriptionsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/subscriptions", handlers.TaskHandler(addUserSubscriptionTask)).Methods("POST").MatcherFunc(addUserSubscriptionTask.Matcher())
	ar.HandleFunc("/user/{user}/subscriptions", handlers.TaskHandler(updateUserSubscriptionTask)).Methods("POST").MatcherFunc(updateUserSubscriptionTask.Matcher())
	ar.HandleFunc("/user/{user}/subscriptions", handlers.TaskHandler(deleteUserSubscriptionTask)).Methods("POST").MatcherFunc(deleteUserSubscriptionTask.Matcher())
	ar.HandleFunc("/user/{user}/comment", adminUserAddCommentPage).Methods("POST")
	ar.HandleFunc("/user/{user}/grants", adminUserGrantsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/grant/add", adminUserGrantAddPage).Methods("GET")
	ar.HandleFunc("/user/{user}/grant", handlers.TaskHandler(userGrantCreateTask)).Methods("POST").MatcherFunc(userGrantCreateTask.Matcher())
	ar.HandleFunc("/user/{user}/grant/update", handlers.TaskHandler(userGrantUpdateTask)).Methods("POST").MatcherFunc(userGrantUpdateTask.Matcher())
	ar.HandleFunc("/role/{role}", adminRolePage).Methods("GET")
	ar.HandleFunc("/role/{role}/edit", adminRoleEditFormPage).Methods("GET")
	ar.HandleFunc("/role/{role}/edit", adminRoleEditSavePage).Methods("POST")
	ar.HandleFunc("/role/{role}/grant/add", adminRoleGrantAddPage).Methods("GET")
	ar.HandleFunc("/role/{role}/grant", handlers.TaskHandler(roleGrantCreateTask)).Methods("POST").MatcherFunc(roleGrantCreateTask.Matcher())
	ar.HandleFunc("/role/{role}/grant/update", handlers.TaskHandler(roleGrantUpdateTask)).Methods("POST").MatcherFunc(roleGrantUpdateTask.Matcher())
	ar.HandleFunc("/grant/delete", handlers.TaskHandler(roleGrantDeleteTask)).Methods("POST").MatcherFunc(roleGrantDeleteTask.Matcher())
	ar.HandleFunc("/maintenance", AdminMaintenancePage).Methods("GET")
	ar.HandleFunc("/maintenance", handlers.TaskHandler(convertTopicToPrivateTask)).Methods("POST").MatcherFunc(convertTopicToPrivateTask.Matcher())
	ar.HandleFunc("/grants/anyone", AdminAnyoneGrantsPage).Methods("GET")
	ar.HandleFunc("/grants/available", AdminGrantsAvailablePage).Methods("GET")
	ar.HandleFunc("/grants", AdminGrantsPage).Methods("GET")
	ar.HandleFunc("/grant/add", adminGrantAddPage).Methods("GET")
	ar.HandleFunc("/grant/bulk", handlers.TaskHandler(grantBulkCreateTask)).Methods("POST").MatcherFunc(grantBulkCreateTask.Matcher())
	ar.HandleFunc("/grant", handlers.TaskHandler(globalGrantCreateTask)).Methods("POST").MatcherFunc(globalGrantCreateTask.Matcher())
	ar.HandleFunc("/grant/{grant}", adminGrantPage).Methods("GET")
	ar.HandleFunc("/grant/update", handlers.TaskHandler(grantUpdateTask)).Methods("POST").MatcherFunc(grantUpdateTask.Matcher())
	ar.HandleFunc("/user/{user}/reset", adminUserResetPasswordConfirmPage).Methods("GET")
	ar.HandleFunc("/user/{user}/reset", handlers.TaskHandler(userForcePasswordChangeTask)).Methods("POST").MatcherFunc(userForcePasswordChangeTask.Matcher())
	ar.HandleFunc("/user/{user}/reset", handlers.TaskHandler(userSendResetEmailTask)).Methods("POST").MatcherFunc(userSendResetEmailTask.Matcher())
	ar.HandleFunc("/user/{user}/reset", handlers.TaskHandler(userGenerateResetLinkTask)).Methods("POST").MatcherFunc(userGenerateResetLinkTask.Matcher())
	ar.HandleFunc("/announcements", AdminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", handlers.TaskHandler(addAnnouncementTask)).Methods("POST").MatcherFunc(addAnnouncementTask.Matcher())
	ar.HandleFunc("/announcements", handlers.TaskHandler(deleteAnnouncementTask)).Methods("POST").MatcherFunc(deleteAnnouncementTask.Matcher())
	ar.HandleFunc("/comments", AdminCommentsPage).Methods("GET")
	ar.HandleFunc("/comments/deactivated", AdminDeactivatedCommentsPage).Methods("GET")
	ar.HandleFunc("/comment/{comment}", adminCommentPage).Methods("GET")
	ar.HandleFunc("/comment/{comment}", handlers.TaskHandler(editCommentTask)).Methods("POST").MatcherFunc(editCommentTask.Matcher())
	ar.HandleFunc("/comment/{comment}", handlers.TaskHandler(deleteCommentTask)).Methods("POST").MatcherFunc(deleteCommentTask.Matcher())
	ar.HandleFunc("/comment/{comment}", handlers.TaskHandler(deactivateCommentTask)).Methods("POST").MatcherFunc(deactivateCommentTask.Matcher())
	ar.HandleFunc("/comment/{comment}", handlers.TaskHandler(restoreCommentTask)).Methods("POST").MatcherFunc(restoreCommentTask.Matcher())
	ar.HandleFunc("/ipbans", AdminIPBanPage).Methods("GET")
	ar.HandleFunc("/ipbans/export", AdminIPBanExport).Methods("GET")
	ar.HandleFunc("/ipbans", handlers.TaskHandler(addIPBanTask)).Methods("POST").MatcherFunc(addIPBanTask.Matcher())
	ar.HandleFunc("/ipbans", handlers.TaskHandler(deleteIPBanTask)).Methods("POST").MatcherFunc(deleteIPBanTask.Matcher())
	ar.HandleFunc("/ipbans", handlers.TaskHandler(ipBanBulkTask)).Methods("POST").MatcherFunc(ipBanBulkTask.Matcher())
	ar.HandleFunc("/audit", AdminAuditLogPage).Methods("GET")
	ar.HandleFunc("/settings", h.AdminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/config/as-cli", h.AdminConfigAsCLIPage).Methods("GET")
	ar.HandleFunc("/config/explain", h.AdminConfigExplainPage).Methods("GET")
	ar.HandleFunc("/page-size", AdminPageSizePage).Methods("GET", "POST")
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/files/unmanaged", AdminUnmanagedFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	deleteUnmanagedFileTask := h.NewDeleteUnmanagedFileTask()
	ar.HandleFunc("/files/delete", handlers.TaskHandler(deleteUnmanagedFileTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(deleteUnmanagedFileTask.Matcher())
	ar.HandleFunc("/images/cache", AdminImageCachePage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/images/cache", handlers.TaskHandler(imageCacheListTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(imageCacheListTask.Matcher())
	ar.HandleFunc("/images/cache", handlers.TaskHandler(imageCachePruneTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(imageCachePruneTask.Matcher())
	ar.HandleFunc("/images/cache", handlers.TaskHandler(imageCacheDeleteTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(imageCacheDeleteTask.Matcher())
	ar.HandleFunc("/images/cache/{id}", AdminImageCacheDetailsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	ar.HandleFunc("/images/cache/{id}", handlers.TaskHandler(imageCacheRefreshTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(imageCacheRefreshTask.Matcher())
	ar.HandleFunc("/share/tools", AdminShareToolsPage).Methods("GET", "POST")
	ar.HandleFunc("/stats", h.AdminServerStatsPage).Methods("GET")
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
	news.RegisterAdminRoutes(ar)

	// writings admin
	writings.RegisterAdminRoutes(ar)

	// Verify administrator access within the handlers so direct CLI calls
	// cannot bypass the permission checks.
	ar.HandleFunc("/reload",
		handlers.RequireRole(h.AdminReloadConfigPage, fmt.Errorf("administrator role required"), "administrator")).
		Methods("POST").
		MatcherFunc(handlers.RequiredAccess("administrator"))
	sst := h.NewServerShutdownTask()
	ar.HandleFunc("/shutdown",
		handlers.RequireRole(handlers.TaskHandler(sst), fmt.Errorf("administrator role required"), "administrator")).
		Methods("POST").
		MatcherFunc(handlers.RequiredAccess("administrator")).
		MatcherFunc(sst.Matcher())

	api := ar.PathPrefix("/api").Subrouter()
	api.Use(router.AdminCheckerMiddleware)
	api.HandleFunc("/shutdown", h.AdminAPIServerShutdown).MatcherFunc(AdminAPISigned()).Methods("POST")
}

// Register registers the admin router module.
