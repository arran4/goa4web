package db

import (
	"context"
	"database/sql"
)

// SessionProxy forwards session operations to Queries while satisfying
// common.SessionManager without exposing database types.
type SessionProxy struct {
	*Queries
}

// NewSessionProxy returns a SessionProxy wrapping q.
func NewSessionProxy(q *Queries) *SessionProxy {
	return &SessionProxy{Queries: q}
}

// InsertSession stores or updates a session record.
func (sp *SessionProxy) InsertSession(ctx context.Context, sessionID string, userID int32, branchName string) error {
	return sp.Queries.SystemInsertSession(ctx, SystemInsertSessionParams{
		SessionID:    sessionID,
		UsersIdusers: userID,
		BranchName:   sql.NullString{String: branchName, Valid: branchName != ""},
	})
}

// DeleteSessionByID removes a session record by ID.
func (sp *SessionProxy) DeleteSessionByID(ctx context.Context, sessionID string) error {
	return sp.Queries.SystemDeleteSessionByID(ctx, sessionID)
}
