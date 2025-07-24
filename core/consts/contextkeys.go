package consts

// ContextKey is used for storing values in the request context.
type ContextKey string

// Context keys used across the handler packages.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData ContextKey = "coreData"
	// KeyTopic holds the current topic information.
	KeyTopic ContextKey = "topic"
	// KeyBlogEntry holds a fetched blog entry row.
	KeyBlogEntry ContextKey = "blogEntry"
	// KeyNewsPost holds the news post row.
	KeyNewsPost ContextKey = "newsPost"
	// KeyWriting contains the writing row.
	KeyWriting ContextKey = "writing"
)
