package handlers

import common "github.com/arran4/goa4web/core/common"

// Common context keys used across handlers.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData common.ContextValues = "coreData"
	// KeyQueries holds the db.Queries pointer.
	KeyQueries common.ContextValues = "queries"
	// KeySQLDB exposes the *sql.DB handle.
	KeySQLDB common.ContextValues = "sql.DB"
	// KeyThread holds the current thread information.
	KeyThread common.ContextValues = "thread"
	// KeyTopic holds the current topic information.
	KeyTopic common.ContextValues = "topic"
	// KeyBlogEntry holds a fetched blog entry row.
	KeyBlogEntry common.ContextValues = "blogEntry"
	// KeyComment stores the current comment row.
	KeyComment common.ContextValues = "comment"
	// KeyNewsPost holds the news post row.
	KeyNewsPost common.ContextValues = "newsPost"
	// KeyWriting contains the writing row.
	KeyWriting common.ContextValues = "writing"
)
