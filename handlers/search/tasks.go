package search

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// These strings correspond to form "task" values and matcher checks in routes.
const (
	// TaskSearchForum triggers a forum search.
	TaskSearchForum tasks.TaskString = "Search forum"

	// TaskSearchNews triggers a news search.
	TaskSearchNews tasks.TaskString = "Search news"

	// TaskSearchLinker triggers a linker search.
	TaskSearchLinker tasks.TaskString = "Search linker"

	// TaskSearchBlogs triggers a blog search.
	TaskSearchBlogs tasks.TaskString = "Search blogs"

	// TaskSearchWritings triggers a writing search.
	TaskSearchWritings tasks.TaskString = "Search writings"

	// TaskRemakeCommentsSearch rebuilds the comments search index.
	TaskRemakeCommentsSearch tasks.TaskString = "Remake comments search"

	// TaskRemakeNewsSearch rebuilds the news search index.
	TaskRemakeNewsSearch tasks.TaskString = "Remake news search"

	// TaskRemakeBlogSearch rebuilds the blog search index.
	TaskRemakeBlogSearch tasks.TaskString = "Remake blog search"

	// TaskRemakeLinkerSearch rebuilds the linker search index.
	TaskRemakeLinkerSearch tasks.TaskString = "Remake linker search"

	// TaskRemakeWritingSearch rebuilds the writing search index.
	TaskRemakeWritingSearch tasks.TaskString = "Remake writing search"

	// TaskRemakeImageSearch rebuilds the image search index.
	TaskRemakeImageSearch tasks.TaskString = "Remake image search"
)

const (
	// TaskRemakeCommentsSearchComplete notifies comment index rebuild completion.
	TaskRemakeCommentsSearchComplete tasks.TaskString = "Remake comments search complete"
	// TaskRemakeNewsSearchComplete notifies news index rebuild completion.
	TaskRemakeNewsSearchComplete tasks.TaskString = "Remake news search complete"
	// TaskRemakeBlogSearchComplete notifies blog index rebuild completion.
	TaskRemakeBlogSearchComplete tasks.TaskString = "Remake blog search complete"
	// TaskRemakeLinkerSearchComplete notifies linker index rebuild completion.
	TaskRemakeLinkerSearchComplete tasks.TaskString = "Remake linker search complete"
	// TaskRemakeWritingSearchComplete notifies writing index rebuild completion.
	TaskRemakeWritingSearchComplete tasks.TaskString = "Remake writing search complete"
	// TaskRemakeImageSearchComplete notifies image index rebuild completion.
	TaskRemakeImageSearchComplete tasks.TaskString = "Remake image search complete"
)
