package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminTemplatesExist(t *testing.T) {
	pages := []tasks.Template{
		AdminAuditLogPageTmpl,
		AdminCommentPageTmpl,
		AdminLinkRemapPageTmpl,
		AdminGrantAddPageTmpl,
		AdminRoleGrantAddPageTmpl,
		AdminPageTmpl,
		AdminUserListPageTmpl,
		AdminUserCommentsPageTmpl,
		AdminUserLinkerPageTmpl,
		AdminCommentsPageTmpl,
		AdminEmailTemplateListPageTmpl,
		AdminEmailTemplateEditPageTmpl,
		AdminTemplateExportPageTmpl,
		AdminEmailPageTmpl,
		AdminNotificationsPageTmpl,
		AdminServerStatsPageTmpl,
		AdminRolesPageTmpl,
		AdminRoleLoadPageTmpl,
		AdminRoleTemplatesPageTmpl,
		AdminUsageStatsPageTmpl,
		AdminUserBlogsPageTmpl,
		AdminMaintenancePageTmpl,
		AdminDBStatusPageTmpl,
		AdminDBSchemaPageTmpl,
		AdminDBMigrationsPageTmpl,
		AdminSiteSettingsPageTmpl,
		AdminConfigExplainPageTmpl,
		AdminPageSizePageTmpl,
		AdminDeactivatedCommentsPageTmpl,
		TemplateUserResetPasswordConfirmPage, // Note: This constant name was kept as is in adminUserPasswordReset.go
		AdminAnnouncementsPageTmpl,
		AdminEmailTestPageTmpl,
		AdminUserWritingsPageTmpl,
		AdminRolePageTmpl,
		AdminFilesPageTmpl,
		AdminImageCachePageTmpl,
		AdminRequestQueuePageTmpl,
		AdminRequestArchivePageTmpl,
		AdminRequestPageTmpl,
		AdminGrantsAvailablePageTmpl,
		AdminGrantsPageTmpl,
		GrantPageTmpl,
		AdminUserGrantsPageTmpl,
		AdminUserGrantAddPageTmpl,
		AdminUserProfilePageTmpl,
		AdminRoleEditPageTmpl,
		AdminDLQPageTmpl,
		AdminUserSubscriptionsPageTmpl,
		AdminExternalLinksPageTmpl,
		AdminLinksToolsPageTmpl,
		AdminUserForumPageTmpl,
		AdminUserImagebbsPageTmpl,
		RunTaskPageTmpl, // Common template used in multiple places
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(), "Page %s should exist", page)
		})
	}
}
