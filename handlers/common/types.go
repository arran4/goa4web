package common

import common "github.com/arran4/goa4web/core/common"

// ContextKey maps to core.ContextValues.
type ContextKey = common.ContextValues

// CoreData maps to core.CoreData for handler use.
type CoreData = common.CoreData

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
	// KeyBlogEntry holds a fetched blog entry row.
	KeyBlogEntry ContextKey = "blogEntry"
	// KeyComment stores the current comment row.
	KeyComment ContextKey = "comment"
	// KeyNewsPost holds the news post row.
	KeyNewsPost ContextKey = "newsPost"
	// KeyWriting contains the writing row.
	KeyWriting ContextKey = "writing"
	// KeyBusEvent stores the pointer to the event being built by middleware.
	KeyBusEvent ContextKey = "busEvent"
)
