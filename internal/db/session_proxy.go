package db

import "context"

// SessionProxy forwards session operations to Queries while satisfying
// common.SessionManager without exposing database types.
type SessionProxy struct {
	*Queries
}

// TODO coredata should provide this this hsould be injected into server & coredata as part of DI
// NewSessionProxy returns a SessionProxy wrapping q.
func NewSessionProxy(q *Queries) *SessionProxy {
	return &SessionProxy{Queries: q}
}

// InsertSession stores or updates a session record.
func (sp *SessionProxy) InsertSession(ctx context.Context, sessionID string, userID int32) error {
	return sp.Queries.SystemInsertSession(ctx, SystemInsertSessionParams{SessionID: sessionID, UsersIdusers: userID})
}

// DeleteSessionByID removes a session record by ID.
func (sp *SessionProxy) DeleteSessionByID(ctx context.Context, sessionID string) error {
	return sp.Queries.SystemDeleteSessionByID(ctx, sessionID)
}
