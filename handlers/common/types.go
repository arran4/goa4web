package common

import "github.com/arran4/goa4web/core"

// ContextKey maps to core.ContextValues.
type ContextKey = core.ContextValues
type ContextValues = core.ContextValues

// CoreData maps to core.CoreData for handler use.
type CoreData = core.CoreData

// Common context keys used across handlers.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData ContextKey = "coreData"
	// KeyLanguages provides the user's language preferences.
	KeyLanguages ContextKey = "languages"
	// KeyPermissions stores user permissions.
	KeyPermissions ContextKey = "permissions"
	// KeyPreference stores a user's preferences.
	KeyPreference ContextKey = "preference"
	// KeyQueries holds the db.Queries pointer.
	KeyQueries ContextKey = "queries"
	// KeySession contains the session struct.
	KeySession ContextKey = "session"
	// KeySQLDB exposes the *sql.DB handle.
	KeySQLDB ContextKey = "sql.DB"
	// KeyThread holds the current thread information.
	KeyThread ContextKey = "thread"
	// KeyTopic holds the current topic information.
	KeyTopic ContextKey = "topic"
	// KeyUser references the authenticated user.
	KeyUser ContextKey = "user"
)
