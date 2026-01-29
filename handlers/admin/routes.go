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
)

// RegisterRoutes attaches the admin endpoints to ar. The router is expected to
// already have any required authentication middleware applied.
func (h *Handlers) RegisterRoutes(ar *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
	ar.Use(handlers.SectionMiddleware("admin"))
	navReg.RegisterAdminControlCenter("Core", "Roles", "/admin/roles", 25)
	navReg.RegisterAdminControlCenter("Core", "Grants", "/admin/grants", 27)
	navReg.RegisterAdminControlCenter("Core", "Available Grants", "/admin/grants/available", 28)
	navReg.RegisterAdminControlCenter("Core", "Once Off & Maintenance", "/admin/maintenance", 29)
	navReg.RegisterAdminControlCenter("Core", "External Links", "/admin/external-links", 30)
	navReg.RegisterAdminControlCenter("Core", "Link Remap", "/admin/link-remap", 32)
	navReg.RegisterAdminControlCenter("Core", "Notifications", "/admin/notifications", 90)
	navReg.RegisterAdminControlCenter("Core", "Queued Emails", "/admin/email/queue", 110)
	navReg.RegisterAdminControlCenter("Core", "Failed Emails", "/admin/email/failed", 112)
	navReg.RegisterAdminControlCenter("Core", "Sent Emails", "/admin/email/sent", 115)
	navReg.RegisterAdminControlCenter("Core", "Email Tester", "/admin/email/test", 118)
	navReg.RegisterAdminControlCenter("Core", "Email Template", "/admin/email/template", 120)
	navReg.RegisterAdminControlCenter("Core", "Dead Letter Queue", "/admin/dlq", 130)
	navReg.RegisterAdminControlCenter("Core", "Server Stats", "/admin/stats", 140)
	navReg.RegisterAdminControlCenter("Core", "Requests", "/admin/requests", 145)
	navReg.RegisterAdminControlCenter("Core", "Password Resets", "/admin/password_resets", 146)
	navReg.RegisterAdminControlCenter("Core", "Comments", "/admin/comments", 147)
	navReg.RegisterAdminControlCenter("Core", "Deactivated Comments", "/admin/comments/deactivated", 148)
	navReg.RegisterAdminControlCenter("Core", "Site Settings", "/admin/settings", 150)
	navReg.RegisterAdminControlCenter("Core", "Pagination", "/admin/page-size", 152)
	navReg.RegisterAdminControlCenter("Core", "Files", "/admin/files", 153)
	navReg.RegisterAdminControlCenter("Core", "Usage Stats", "/admin/usage", 160)

	ar.HandleFunc("", AdminPage).Methods("GET")
	ar.HandleFunc("/", AdminPage).Methods("GET")
	ar.HandleFunc("/role-grants-editor.js", handlers.RoleGrantsEditorJS(cfg)).Methods(http.MethodGet, http.MethodHead, http.MethodOptions)
	ar.HandleFunc("/roles", AdminRolesPage).Methods("GET")
	ar.HandleFunc("/roles", handlers.TaskHandler(rolePublicProfileTask)).Methods("POST").MatcherFunc(rolePublicProfileTask.Matcher())
	ar.HandleFunc("/external-links", AdminExternalLinksPage).Methods("GET")
	ar.HandleFunc("/external-links", handlers.TaskHandler(refreshExternalLinkTask)).Methods("POST").MatcherFunc(refreshExternalLinkTask.Matcher())
	ar.HandleFunc("/external-links", handlers.TaskHandler(deleteExternalLinkTask)).Methods("POST").MatcherFunc(deleteExternalLinkTask.Matcher())
	ar.HandleFunc("/link-remap", AdminLinkRemapPage).Methods("GET")
	ar.HandleFunc("/link-remap", handlers.TaskHandler(applyLinkRemapTask)).Methods("POST").MatcherFunc(applyLinkRemapTask.Matcher())
	ar.HandleFunc("/email/queue", AdminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/failed", AdminFailedEmailsPage).Methods("GET")
	ar.HandleFunc("/email/sent", AdminSentEmailsPage).Methods("GET")
	ar.HandleFunc("/email/sent", handlers.TaskHandler(resendSentEmailTask)).Methods("POST").MatcherFunc(resendSentEmailTask.Matcher())
	ar.HandleFunc("/email/sent", handlers.TaskHandler(retrySentEmailTask)).Methods("POST").MatcherFunc(retrySentEmailTask.Matcher())
	ar.HandleFunc("/email/test", h.AdminEmailTestPage).Methods("GET", "POST")
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
	ar.HandleFunc("/request/{request}", adminRequestPage).Methods("GET")
	ar.HandleFunc("/request/{request}/comment", adminRequestAddCommentPage).Methods("POST")
	ar.HandleFunc("/request/{request}/accept", handlers.TaskHandler(acceptRequestTask)).Methods("POST").MatcherFunc(acceptRequestTask.Matcher())
	ar.HandleFunc("/request/{request}/reject", handlers.TaskHandler(rejectRequestTask)).Methods("POST").MatcherFunc(rejectRequestTask.Matcher())
	ar.HandleFunc("/request/{request}/query", handlers.TaskHandler(queryRequestTask)).Methods("POST").MatcherFunc(queryRequestTask.Matcher())
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
	ar.HandleFunc("/page-size", AdminPageSizePage).Methods("GET", "POST")
	ar.HandleFunc("/files", AdminFilesPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
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
		handlers.VerifyAccess(h.AdminReloadConfigPage, fmt.Errorf("administrator role required"), "administrator")).
		Methods("POST").
		MatcherFunc(handlers.RequiredAccess("administrator"))
	sst := h.NewServerShutdownTask()
	ar.HandleFunc("/shutdown",
		handlers.VerifyAccess(handlers.TaskHandler(sst), fmt.Errorf("administrator role required"), "administrator")).
		Methods("POST").
		MatcherFunc(handlers.RequiredAccess("administrator")).
		MatcherFunc(sst.Matcher())

	api := ar.PathPrefix("/api").Subrouter()
	api.Use(router.AdminCheckerMiddleware)
	api.HandleFunc("/shutdown", h.AdminAPIServerShutdown).MatcherFunc(AdminAPISigned()).Methods("POST")
}

// Register registers the admin router module.
