package consts

// ContextKey is used for storing values in the request context.
type ContextKey string

// Context keys used across the handler packages.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData ContextKey = "coreData"
)
