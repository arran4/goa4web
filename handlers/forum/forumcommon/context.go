package forumcommon

// ForumContext encapsulates configuration for a specific forum instance.
type ForumContext struct {
	Section  string
	BasePath string
}

// New creates a new ForumContext.
func New(section, basePath string) *ForumContext {
	return &ForumContext{
		Section:  section,
		BasePath: basePath,
	}
}
