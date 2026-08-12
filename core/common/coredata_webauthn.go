package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/db"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"encoding/gob"
)

func init() {
	gob.Register(webauthn.SessionData{})
}

type WebAuthnUser struct {
	Idusers  int32
	Username string
	Passkeys []*db.UserPasskey
}

// WebAuthnID returns the user's ID
func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(fmt.Sprintf("%d", u.Idusers))
}

// WebAuthnName returns the user's username
func (u *WebAuthnUser) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName returns the user's display name
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.Username
}

// WebAuthnIcon is not (yet) implemented
func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns credentials owned by the user
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	var creds []webauthn.Credential
	for _, p := range u.Passkeys {
		creds = append(creds, webauthn.Credential{
			ID:              p.CredentialID,
			PublicKey:       p.PublicKey,
			AttestationType: p.AttestationType,
			Flags: webauthn.CredentialFlags{
				BackupEligible: p.BackupEligible.Valid && p.BackupEligible.Bool,
				BackupState:    p.BackupState.Valid && p.BackupState.Bool,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:       p.Aaguid,
				SignCount:    uint32(p.SignCount),
				CloneWarning: false,
			},
		})
	}
	return creds
}

// EstablishLegacyCredentialFlags uses the current assertion flags only for a
// credential created before backup flags were persisted. Validation still
// performs the normal WebAuthn backup-eligibility consistency check.
func (u *WebAuthnUser) EstablishLegacyCredentialFlags(credentialID []byte, flags protocol.AuthenticatorFlags) {
	for _, passkey := range u.Passkeys {
		if bytes.Equal(passkey.CredentialID, credentialID) && !passkey.BackupEligible.Valid {
			passkey.BackupEligible = sql.NullBool{Bool: flags.HasBackupEligible(), Valid: true}
			passkey.BackupState = sql.NullBool{Bool: flags.HasBackupState(), Valid: true}
			return
		}
	}
}

// Ensure interface is fulfilled
var _ webauthn.User = (*WebAuthnUser)(nil)

func (cd *CoreData) GetWebAuthnUser(username string) (*WebAuthnUser, error) {
	row, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		return nil, err
	}

	passkeys, err := cd.queries.GetPasskeysByUserID(cd.ctx, row.Idusers)
	if err != nil {
		return nil, err
	}

	return &WebAuthnUser{
		Idusers:  row.Idusers,
		Username: row.Username.String,
		Passkeys: passkeys,
	}, nil
}

func (cd *CoreData) GetWebAuthnUserByID(id int32) (*WebAuthnUser, error) {
	row, err := cd.queries.SystemGetUserByID(cd.ctx, id)
	if err != nil {
		return nil, err
	}

	passkeys, err := cd.queries.GetPasskeysByUserID(cd.ctx, id)
	if err != nil {
		return nil, err
	}

	return &WebAuthnUser{
		Idusers:  row.Idusers,
		Username: row.Username.String,
		Passkeys: passkeys,
	}, nil
}

func (cd *CoreData) SavePasskey(passkey *webauthn.Credential, userID int32, name string) error {
	return cd.queries.InsertPasskey(cd.ctx, db.InsertPasskeyParams{
		UserID:          userID,
		Name:            name,
		BackupEligible:  sql.NullBool{Bool: passkey.Flags.BackupEligible, Valid: true},
		BackupState:     sql.NullBool{Bool: passkey.Flags.BackupState, Valid: true},
		CredentialID:    passkey.ID,
		PublicKey:       passkey.PublicKey,
		AttestationType: passkey.AttestationType,
		Aaguid:          passkey.Authenticator.AAGUID,
		SignCount:       int32(passkey.Authenticator.SignCount),
	})
}

func (cd *CoreData) UpdatePasskeyAfterLogin(cred *webauthn.Credential) error {
	return cd.queries.UpdatePasskeyAfterLogin(cd.ctx, db.UpdatePasskeyAfterLoginParams{
		CredentialID:   cred.ID,
		SignCount:      int32(cred.Authenticator.SignCount),
		BackupEligible: sql.NullBool{Bool: cred.Flags.BackupEligible, Valid: true},
		BackupState:    sql.NullBool{Bool: cred.Flags.BackupState, Valid: true},
	})
}
