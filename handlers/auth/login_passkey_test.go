package auth

import (
	"net/http"
	"net/http/httptest"
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

func TestPasskeyUnavailableResponseDoesNotRevealUserExistence(t *testing.T) {
	t.Run("Missing user and user without passkeys share response", func(t *testing.T) {
		responses := make([]*httptest.ResponseRecorder, 2)
		for i := range responses {
			responses[i] = httptest.NewRecorder()
			writePasskeyUnavailable(responses[i])
		}

		if responses[0].Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", responses[0].Code, http.StatusNotFound)
		}
		if responses[0].Code != responses[1].Code || responses[0].Body.String() != responses[1].Body.String() {
			t.Fatal("passkey availability responses differ")
		}
	})
}
