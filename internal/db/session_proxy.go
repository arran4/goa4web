package db

import "context"

// SessionProxy forwards session operations to Queries while satisfying
// common.SessionManager without exposing database types.
type SessionProxy struct {
	Querier
}

// NewSessionProxy returns a SessionProxy wrapping q.
func NewSessionProxy(q Querier) *SessionProxy {
	return &SessionProxy{Querier: q}
}

// InsertSession stores or updates a session record.
func (sp *SessionProxy) InsertSession(ctx context.Context, sessionID string, userID int32) error {
	return sp.Querier.SystemInsertSession(ctx, SystemInsertSessionParams{SessionID: sessionID, UsersIdusers: userID})
}

// DeleteSessionByID removes a session record by ID.
func (sp *SessionProxy) DeleteSessionByID(ctx context.Context, sessionID string) error {
	return sp.Querier.SystemDeleteSessionByID(ctx, sessionID)
}
