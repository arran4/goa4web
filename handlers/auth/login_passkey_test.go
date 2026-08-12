package auth

import (
	"strings"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/arran4/goa4web/core/common"
)

func TestPasskeyLoginUsesUserBoundCeremony(t *testing.T) {
	t.Run("BeginLogin session is validated as a known user", func(t *testing.T) {
		user := &common.WebAuthnUser{Idusers: 1, Username: "user"}
		session := webauthn.SessionData{UserID: []byte("different-user")}

		_, err := validatePasskeyLogin(&webauthn.WebAuthn{}, user, session, &protocol.ParsedCredentialAssertionData{})
		if err == nil {
			t.Fatal("ValidateLogin returned no error for mismatched user IDs")
		}
		if !strings.Contains(err.Error(), "ID mismatch") {
			t.Fatalf("ValidateLogin error = %q, want an ID mismatch", err)
		}
		if strings.Contains(err.Error(), "discoverable login") {
			t.Fatalf("known-user login was incorrectly validated as discoverable: %v", err)
		}
	})
}
