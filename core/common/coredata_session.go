package common

import (
	bm "github.com/arran4/gobookmarks/core"
	"github.com/gorilla/sessions"
)

// This satisfies the gobookmarks.Core interface.
func (cd *CoreData) GetSession() *sessions.Session {
	return cd.session
}

// GetUser returns the current user.
// This satisfies the gobookmarks.Core interface.
func (cd *CoreData) GetUser() bm.User {
	if v, ok := cd.user.Peek(); ok && v != nil {
		return v
	}
	return nil
}
