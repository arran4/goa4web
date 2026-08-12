package common

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/arran4/goa4web/internal/db"
)

func TestBackupEligiblePasskeyPersistence(t *testing.T) {
	t.Run("Happy Path - registration and reconstruction preserve flags", func(t *testing.T) {
		var inserted db.InsertPasskeyParams
		queries := &db.QuerierStub{InsertPasskeyStub: func(_ context.Context, arg db.InsertPasskeyParams) error {
			inserted = arg
			return nil
		}}
		cd := NewCoreData(context.Background(), queries, nil)
		credential := &webauthn.Credential{
			ID:        []byte("credential"),
			PublicKey: []byte("public-key"),
			Flags: webauthn.CredentialFlags{
				BackupEligible: true,
				BackupState:    true,
			},
		}

		if err := cd.SavePasskey(credential, 1, "Phone"); err != nil {
			t.Fatalf("SavePasskey: %v", err)
		}
		if !inserted.BackupEligible.Valid || !inserted.BackupEligible.Bool {
			t.Fatalf("persisted BackupEligible = %+v, want true", inserted.BackupEligible)
		}
		if !inserted.BackupState.Valid || !inserted.BackupState.Bool {
			t.Fatalf("persisted BackupState = %+v, want true", inserted.BackupState)
		}

		user := &WebAuthnUser{Passkeys: []*db.UserPasskey{{
			CredentialID:   inserted.CredentialID,
			PublicKey:      inserted.PublicKey,
			BackupEligible: inserted.BackupEligible,
			BackupState:    inserted.BackupState,
		}}}
		reconstructed := user.WebAuthnCredentials()
		if len(reconstructed) != 1 || !reconstructed[0].Flags.BackupEligible || !reconstructed[0].Flags.BackupState {
			t.Fatalf("reconstructed flags = %+v, want backup eligible and backed up", reconstructed)
		}
		assertionFlags := protocol.FlagBackupEligible | protocol.FlagBackupState
		if reconstructed[0].Flags.BackupEligible != assertionFlags.HasBackupEligible() {
			t.Fatal("reconstructed credential would fail WebAuthn backup eligibility consistency validation")
		}
	})

	t.Run("Legacy eligibility is not assumed false", func(t *testing.T) {
		passkey := &db.UserPasskey{CredentialID: []byte("legacy"), BackupEligible: sql.NullBool{}}
		user := &WebAuthnUser{Passkeys: []*db.UserPasskey{passkey}}

		user.EstablishLegacyCredentialFlags([]byte("legacy"), protocol.FlagBackupEligible|protocol.FlagBackupState)

		if !passkey.BackupEligible.Valid || !passkey.BackupEligible.Bool || !passkey.BackupState.Bool {
			t.Fatalf("legacy flags were not established from assertion: %+v", passkey)
		}
	})

	t.Run("Login updates state without replacing established eligibility", func(t *testing.T) {
		var updated db.UpdatePasskeyAfterLoginParams
		queries := &db.QuerierStub{UpdatePasskeyAfterLoginStub: func(_ context.Context, arg db.UpdatePasskeyAfterLoginParams) error {
			updated = arg
			return nil
		}}
		cd := NewCoreData(context.Background(), queries, nil)
		credential := &webauthn.Credential{ID: []byte("credential"), Flags: webauthn.CredentialFlags{BackupEligible: true, BackupState: true}}

		if err := cd.UpdatePasskeyAfterLogin(credential); err != nil {
			t.Fatalf("UpdatePasskeyAfterLogin: %v", err)
		}
		if !updated.BackupEligible.Valid || !updated.BackupEligible.Bool || !updated.BackupState.Bool {
			t.Fatalf("updated flags = %+v", updated)
		}
	})
}
