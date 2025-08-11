package templates

// TopicLabel represents a label attached to a thread or content item.
// Type is one of "public", "author", or "private".
type TopicLabel struct {
	Name string
	Type string
}
