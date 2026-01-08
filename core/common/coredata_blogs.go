package common

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
)

// BlogPost returns the currently requested blog entry.
func (cd *CoreData) BlogPost(ops ...lazy.Option[*db.GetBlogEntryForListerByIDRow]) (*db.GetBlogEntryForListerByIDRow, error) {
	return cd.CurrentBlog(ops...)
}

// BlogComments returns comments for the current blog.
func (cd *CoreData) BlogComments() ([]*db.GetCommentsByThreadIdForUserRow, error) {
	if _, err := cd.BlogCommentThread(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return cd.SelectedSectionThreadComments()
}

// BlogCategories returns bloggers as categories (placeholder).
func (cd *CoreData) BlogCategories(r *http.Request) ([]*db.ListBloggersForListerRow, error) {
	return cd.Bloggers(r)
}

// EditableBlogPost returns the blog entry if the current user can edit it.
func (cd *CoreData) EditableBlogPost(id int32) (*db.GetBlogEntryForListerByIDRow, error) {
	blog, err := cd.BlogEntryByID(id)
	if err != nil {
		return nil, err
	}
	if !cd.CanEditBlog(blog.Idblogs, blog.UsersIdusers) {
		return nil, sql.ErrNoRows
	}
	return blog, nil
}

// BlogCommentThread returns the thread associated with the current blog
// and ensures it is selected for comment helpers.
func (cd *CoreData) BlogCommentThread(ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	blog, err := cd.BlogPost()
	if err != nil {
		return nil, err
	}
	if blog == nil || !blog.ForumthreadID.Valid {
		return nil, sql.ErrNoRows
	}
	cd.currentSection = "blogs"
	cd.SetCurrentThreadAndTopic(blog.ForumthreadID.Int32, 0)
	return cd.SelectedThread(ops...)
}

// CreateBlogReply adds a comment reply to the specified blog post.
func (cd *CoreData) CreateBlogReply(commenterID, threadID, entryID, languageID int32, text string) (int64, error) {
	return cd.CreateBlogCommentForCommenter(commenterID, threadID, entryID, languageID, text)
}

// UpdateBlogReply updates an existing blog comment.
func (cd *CoreData) UpdateBlogReply(commentID, commenterID, languageID int32, text string) error {
	if cd.queries == nil {
		return nil
	}
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return fmt.Errorf("parse images: %w", err)
	}
	comment, err := cd.CommentByID(commentID)
	if err != nil || comment == nil {
		return fmt.Errorf("load comment: %w", err)
	}
	if err := cd.validateImagePathsForThread(commenterID, comment.ForumthreadID, paths); err != nil {
		return fmt.Errorf("validate images: %w", err)
	}
	if err := cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   commentID,
		CommenterID: commenterID,
		EditorID:    sql.NullInt32{Int32: commenterID, Valid: commenterID != 0},
	}); err != nil {
		return err
	}
	if err := cd.recordThreadImages(comment.ForumthreadID, paths); err != nil {
		log.Printf("record thread images: %v", err)
	}
	return nil
}

// BloggerProfile loads a blogger by username and stores the user ID.
func (cd *CoreData) BloggerProfile(username string) (*db.SystemGetUserByUsernameRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	row, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: username != ""})
	if err != nil {
		return nil, err
	}
	cd.currentProfileUserID = row.Idusers
	return row, nil
}

// BloggerPosts returns blog posts for the selected blogger.
func (cd *CoreData) BloggerPosts() ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
	return cd.BlogListForSelectedAuthor()
}

// AllBlogs returns blog posts visible to the current user.
func (cd *CoreData) AllBlogs() ([]*db.ListBlogEntriesForListerRow, error) {
	return cd.BlogList()
}
