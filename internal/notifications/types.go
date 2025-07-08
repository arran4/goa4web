package notifications

import "github.com/arran4/goa4web/internal/db"

// ForumReplyInfo represents key details about a forum reply event.
type ForumReplyInfo struct {
	TopicTitle string
	ThreadID   int32
	Thread     *db.GetThreadLastPosterAndPermsRow
}

// ThreadInfo represents a newly created forum thread.
type ThreadInfo struct {
	TopicTitle string
	Author     string
}

// BlogPostInfo represents a new blog post.
type BlogPostInfo struct {
	Title  string
	Author string
}

// WritingInfo represents a new writing submission.
type WritingInfo struct {
	Title  string
	Author string
}

// SignupInfo holds details about a new user registration.
type SignupInfo struct {
	Username string
}
