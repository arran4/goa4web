package user_test

import (
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/user"
)

var allPages = []handlers.Page{
	user.UserPageTmpl,
	user.UserPublicProfileSettingsPage,
	user.AppearancePage,
	user.PublicProfilePage,
	user.UserEmailPage,
	user.UserEmailVerifiedPage,
	user.UserEmailVerifyConfirmPage,
	user.AdminLoginAttemptsPage,
	user.AdminPendingUsersPage,
	user.AdminUserPermissionsPage,
	user.AdminSessionsPage,
	user.AdminUsersPage,
	user.AdminConfirmPage,
	user.AdminRunTaskPage,
	user.AdminUserEditPage,
	user.AdminUserResetPasswordPage,
	user.UserGalleryPage,
	user.UserLangPage,
	user.UserLogoutPage,
	user.UserNotificationsPage,
	user.UserNotificationOpenPage,
	user.UserPagingPage,
	user.UserSubscriptionAddPage,
	user.UserSubscriptionsPage,
	user.UserThreadSubscriptionsPage,
	user.UserTimezonePage,
}

func TestAllRegisteredPagesExist(t *testing.T) {
	for _, p := range allPages {
		if !p.Exists() {
			t.Errorf("Page template missing: %s", p)
		}
	}
}
