package user

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestCustomIndexPasskeys(t *testing.T) {
	t.Run("Happy Path - WebAuthn enabled", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/usr/passkeys", nil)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
		cd.UserID = 1
		cd.WebAuthn = &webauthn.WebAuthn{}

		CustomIndex(cd, req)

		if !common.ContainsItem(cd.CustomIndexItems, "Passkeys") {
			t.Error("passkeys setting is missing from the user index")
		}
	})

	t.Run("WebAuthn disabled", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/usr", nil)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
		cd.UserID = 1

		CustomIndex(cd, req)

		if common.ContainsItem(cd.CustomIndexItems, "Passkeys") {
			t.Error("passkeys setting is visible while WebAuthn is disabled")
		}
	})
}

func TestLogoutClearsUserCustomIndex(t *testing.T) {
	t.Run("Logged-out users cannot see account settings", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
		cd.UserID = 1
		CustomIndex(cd, httptest.NewRequest("GET", "/usr/logout", nil))
		if len(cd.CustomIndexItems) == 0 {
			t.Fatal("test setup did not create account settings")
		}

		clearLoggedOutCoreData(cd)

		if cd.UserID != 0 || len(cd.CustomIndexItems) != 0 {
			t.Fatalf("logout retained account access: user=%d items=%v", cd.UserID, cd.CustomIndexItems)
		}
	})
}
