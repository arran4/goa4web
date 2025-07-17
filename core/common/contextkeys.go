package common

// Context keys used across the handler packages.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData ContextValues = "coreData"
	// KeyQueries holds the db.Queries pointer.
	KeyQueries ContextValues = "queries"
	// KeySQLDB exposes the *sql.DB handle.
	KeySQLDB ContextValues = "sql.DB"
	// KeyThread holds the current thread information.
	KeyThread ContextValues = "thread"
	// KeyTopic holds the current topic information.
	KeyTopic ContextValues = "topic"
	// KeyBlogEntry holds a fetched blog entry row.
	KeyBlogEntry ContextValues = "blogEntry"
	// KeyComment stores the current comment row.
	KeyComment ContextValues = "comment"
	// KeyNewsPost holds the news post row.
	KeyNewsPost ContextValues = "newsPost"
	// KeyWriting contains the writing row.
	KeyWriting ContextValues = "writing"
)
