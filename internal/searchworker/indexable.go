package searchworker

// IndexableTask identifies an event that should be added to the search index.
// Implementations return the index type and extract an identifier from the
// event data map.
type IndexableTask interface {
	// IndexType returns the search index table to update.
	IndexType() string
	// IndexID extracts the record identifier from the event data.
	IndexID(data map[string]any) int64
	// IndexText returns the text that should be indexed.
	IndexText(data map[string]any) string
}
