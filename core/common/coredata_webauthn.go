package common

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/db"
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
			Authenticator: webauthn.Authenticator{
				AAGUID:       p.Aaguid,
				SignCount:    uint32(p.SignCount),
				CloneWarning: false,
			},
		})
	}
	return creds
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
		CredentialID:    passkey.ID,
		PublicKey:       passkey.PublicKey,
		AttestationType: passkey.AttestationType,
		Aaguid:          passkey.Authenticator.AAGUID,
		SignCount:       int32(passkey.Authenticator.SignCount),
	})
}

func ToUpdatePasskeySignCountParams(cred *webauthn.Credential) db.UpdatePasskeySignCountParams {
	return db.UpdatePasskeySignCountParams{
		CredentialID: cred.ID,
		SignCount:    int32(cred.Authenticator.SignCount),
	}
}
