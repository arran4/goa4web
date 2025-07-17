package handlers

import common "github.com/arran4/goa4web/core/common"

// TODO refactor out
// ContextKey maps to core.ContextValues.
type ContextKey = common.ContextValues

// TODO refactor out
// CoreData maps to core.CoreData for handler use.
type CoreData = common.CoreData

// Common context keys used across handlers.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData ContextKey = "coreData"
	// KeyQueries holds the db.Queries pointer.
	KeyQueries ContextKey = "queries"
	// KeySQLDB exposes the *sql.DB handle.
	KeySQLDB ContextKey = "sql.DB"
	// KeyThread holds the current thread information.
	KeyThread ContextKey = "thread"
	// KeyTopic holds the current topic information.
	KeyTopic ContextKey = "topic"
	// KeyBlogEntry holds a fetched blog entry row.
	KeyBlogEntry ContextKey = "blogEntry"
	// KeyComment stores the current comment row.
	KeyComment ContextKey = "comment"
	// KeyNewsPost holds the news post row.
	KeyNewsPost ContextKey = "newsPost"
	// KeyWriting contains the writing row.
	KeyWriting ContextKey = "writing"
)
