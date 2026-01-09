package common

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// newsTopicName is the default name for the hidden news forum.
const newsTopicName = "A NEWS TOPIC"

// newsTopicDescription describes the hidden news forum.
const newsTopicDescription = "THIS IS A HIDDEN FORUM FOR A NEWS TOPIC"

// ThreadInfo summarises forum thread and topic identifiers.
type ThreadInfo struct {
	ThreadID int32
	TopicID  int32
}

// ThreadInfo ensures a news post has a backing forum thread and topic.
func (cd *CoreData) ThreadInfo(post *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow) (ThreadInfo, error) {
	ti := ThreadInfo{}
	if cd.queries == nil || post == nil {
		return ti, nil
	}
	pt, err := cd.queries.SystemGetForumTopicByTitle(cd.ctx, sql.NullString{String: newsTopicName, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		id, err := cd.queries.CreateForumTopicForPoster(cd.ctx, db.CreateForumTopicForPosterParams{
			ForumcategoryID: 0,
			ForumLang:       post.LanguageID,
			Title:           sql.NullString{String: newsTopicName, Valid: true},
			Description:     sql.NullString{String: newsTopicDescription, Valid: true},
			Handler:         "news",
			Section:         "forum",
			GrantCategoryID: sql.NullInt32{},
			GranteeID:       sql.NullInt32{},
			PosterID:        0,
		})
		if err != nil {
			return ti, fmt.Errorf("create forum topic: %w", err)
		}
		ti.TopicID = int32(id)
	} else if err != nil {
		return ti, fmt.Errorf("find forum topic: %w", err)
	} else {
		ti.TopicID = pt.Idforumtopic
	}
	threadID := post.ForumthreadID
	if threadID == 0 {
		id, err := cd.queries.SystemCreateThread(cd.ctx, ti.TopicID)
		if err != nil {
			return ti, fmt.Errorf("create thread: %w", err)
		}
		threadID = int32(id)
		if err := cd.queries.SystemAssignNewsThreadID(cd.ctx, db.SystemAssignNewsThreadIDParams{ForumthreadID: threadID, Idsitenews: post.Idsitenews}); err != nil {
			return ti, fmt.Errorf("assign news thread: %w", err)
		}
	}
	ti.ThreadID = threadID
	return ti, nil
}

// CreateNewsReply creates a comment for a news post ensuring the thread exists.
func (cd *CoreData) CreateNewsReply(commenterID, postID, languageID int32, text string) (int64, ThreadInfo, error) {
	if cd.queries == nil {
		return 0, ThreadInfo{}, nil
	}
	post, err := cd.queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(cd.ctx, db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: commenterID,
		ID:       postID,
		UserID:   sql.NullInt32{Int32: commenterID, Valid: commenterID != 0},
	})
	if err != nil {
		return 0, ThreadInfo{}, fmt.Errorf("get post: %w", err)
	}
	ti, err := cd.ThreadInfo(post)
	if err != nil {
		return 0, ThreadInfo{}, err
	}
	cid, err := cd.CreateNewsCommentForCommenter(commenterID, ti.ThreadID, postID, languageID, text)
	if err != nil {
		return 0, ThreadInfo{}, err
	}
	return cid, ti, nil
}

// UpdateNewsReply updates an existing comment and returns thread information.
func (cd *CoreData) UpdateNewsReply(commentID, editorID, languageID int32, text string) (ThreadInfo, error) {
	if cd.queries == nil {
		return ThreadInfo{}, nil
	}
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return ThreadInfo{}, fmt.Errorf("parse images: %w", err)
	}
	comment, err := cd.CommentByID(commentID)
	if err != nil {
		return ThreadInfo{}, fmt.Errorf("load comment: %w", err)
	}
	if err := cd.validateImagePathsForThread(editorID, comment.ForumthreadID, paths); err != nil {
		return ThreadInfo{}, fmt.Errorf("validate images: %w", err)
	}
	thread, err := cd.queries.GetThreadLastPosterAndPerms(cd.ctx, db.GetThreadLastPosterAndPermsParams{
		ViewerID:      editorID,
		ThreadID:      comment.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: editorID, Valid: editorID != 0},
	})
	if err != nil {
		return ThreadInfo{}, fmt.Errorf("thread fetch: %w", err)
	}
	if err := cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   commentID,
		CommenterID: editorID,
		EditorID:    sql.NullInt32{Int32: editorID, Valid: editorID != 0},
	}); err != nil {
		return ThreadInfo{}, fmt.Errorf("update comment: %w", err)
	}
	if err := cd.recordThreadImages(comment.ForumthreadID, paths); err != nil {
		log.Printf("record thread images: %v", err)
	}
	return ThreadInfo{ThreadID: thread.Idforumthread, TopicID: thread.ForumtopicIdforumtopic}, nil
}

// UpdateNewsPost modifies an existing news post.
func (cd *CoreData) UpdateNewsPost(postID, languageID, userID int32, text string) error {
	if cd.queries == nil {
		return nil
	}
	if err := cd.validateCodeImagesForUser(userID, text); err != nil {
		return fmt.Errorf("validate images: %w", err)
	}
	return cd.queries.UpdateNewsPostForWriter(cd.ctx, db.UpdateNewsPostForWriterParams{
		PostID:      postID,
		GrantPostID: sql.NullInt32{Int32: postID, Valid: true},
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		News:        sql.NullString{String: text, Valid: true},
		GranteeID:   sql.NullInt32{Int32: userID, Valid: userID != 0},
		WriterID:    userID,
	})
}

// DeleteNewsPost deactivates a news post.
func (cd *CoreData) DeleteNewsPost(postID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.DeactivateNewsPost(cd.ctx, postID)
}

// CreateNewsPost inserts a new news post and grants edit rights to the author.
func (cd *CoreData) CreateNewsPost(languageID, userID int32, text string) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	if err := cd.validateCodeImagesForUser(userID, text); err != nil {
		return 0, fmt.Errorf("validate images: %w", err)
	}
	id, err := cd.queries.CreateNewsPostForWriter(cd.ctx, db.CreateNewsPostForWriterParams{
		LanguageID: sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		News:       sql.NullString{String: text, Valid: true},
		WriterID:   userID,
		Occurred:   sql.NullTime{Time: time.Now().UTC(), Valid: true},
		GranteeID:  sql.NullInt32{Int32: userID, Valid: true},
		Timezone:   sql.NullString{String: cd.Location().String(), Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("create post: %w", err)
	}
	if _, err := cd.queries.AdminCreateGrant(cd.ctx, db.AdminCreateGrantParams{
		UserID:   sql.NullInt32{Int32: userID, Valid: true},
		RoleID:   sql.NullInt32{},
		Section:  "news",
		Item:     sql.NullString{String: "post", Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: int32(id), Valid: true},
		Action:   "edit",
	}); err != nil {
		log.Printf("create grant: %v", err)
	}
	return id, nil
}

// SearchNews finds news posts matching searchwords. Returns flags indicating empty or no results.
func (cd *CoreData) SearchNews(r *http.Request, uid int32) ([]*db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow, bool, bool, error) {
	if cd.queries == nil {
		return nil, false, false, nil
	}
	searchWords := cd.searchWordsFromRequest(r)
	if len(searchWords) == 0 {
		return nil, true, false, nil
	}
	var newsIDs []int32
	for i, word := range searchWords {
		if i == 0 {
			ids, err := cd.queries.ListSiteNewsSearchFirstForLister(cd.ctx, db.ListSiteNewsSearchFirstForListerParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, false, false, fmt.Errorf("news search first: %w", err)
				}
			}
			newsIDs = ids
		} else {
			ids, err := cd.queries.ListSiteNewsSearchNextForLister(cd.ctx, db.ListSiteNewsSearchNextForListerParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ids:      newsIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, false, false, fmt.Errorf("news search next: %w", err)
				}
			}
			newsIDs = ids
		}
		if len(newsIDs) == 0 {
			return nil, false, true, nil
		}
	}
	news, err := cd.queries.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount(cd.ctx, db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		Newsids:  newsIDs,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, false, false, fmt.Errorf("get news: %w", err)
		}
	}
	return news, false, false, nil
}

// AllowNewsUser grants a role to username.
func (cd *CoreData) AllowNewsUser(username, role string) error {
	if cd.queries == nil {
		return nil
	}
	u, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	return cd.queries.SystemCreateUserRole(cd.ctx, db.SystemCreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         role,
	})
}

// DisallowNewsUser removes a role assignment.
func (cd *CoreData) DisallowNewsUser(permissionID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteUserRole(cd.ctx, permissionID)
}

// AddAnnouncement enables or activates an announcement for a news post.
func (cd *CoreData) AddAnnouncement(postID int32) error {
	if cd.queries == nil {
		return nil
	}
	ann, err := cd.NewsAnnouncementWithErr(postID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get announcement: %w", err)
	}
	if ann == nil {
		return cd.queries.AdminPromoteAnnouncement(cd.ctx, postID)
	}
	if !ann.Active {
		return cd.queries.AdminSetAnnouncementActive(cd.ctx, db.AdminSetAnnouncementActiveParams{Active: true, ID: ann.ID})
	}
	return nil
}

// DeleteAnnouncement deactivates an announcement for a news post.
func (cd *CoreData) DeleteAnnouncement(postID int32) error {
	if cd.queries == nil {
		return nil
	}
	ann, err := cd.NewsAnnouncementWithErr(postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("announcement for news: %w", err)
	}
	if ann != nil && ann.Active {
		return cd.queries.AdminSetAnnouncementActive(cd.ctx, db.AdminSetAnnouncementActiveParams{Active: false, ID: ann.ID})
	}
	return nil
}

// SystemGetNewsPost returns a news post by ID without permission checks, iterating through all posts.
// This is inefficient but necessary for shared previews where no specific system query exists.
func (cd *CoreData) SystemGetNewsPost(id int32) (*db.GetAllSiteNewsForIndexRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	// Fetch ALL news posts
	rows, err := cd.queries.GetAllSiteNewsForIndex(cd.ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.Idsitenews == id {
			return row, nil
		}
	}
	return nil, sql.ErrNoRows
}
