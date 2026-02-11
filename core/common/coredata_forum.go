package common

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/internal/db"
)

// ForumCategories loads all forum categories once.
func (cd *CoreData) ForumCategories() ([]*db.Forumcategory, error) {
	return cd.cache.forumCategories.Load(func() ([]*db.Forumcategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetAllForumCategories(cd.ctx, db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	})
}

// ForumCategory loads a single forum category by its identifier.
func (cd *CoreData) ForumCategory(id int32) (*db.Forumcategory, error) {
	if cd.queries == nil {
		return nil, nil
	}
	return cd.queries.GetForumCategoryById(cd.ctx, db.GetForumCategoryByIdParams{Idforumcategory: id, ViewerID: cd.UserID})
}

// ForumThreadByID returns a single forum thread lazily loading it once per ID.
func (cd *CoreData) ForumThreadByID(id int32, ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	fetch := func(i int32) (*db.GetThreadLastPosterAndPermsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetThreadLastPosterAndPerms(cd.ctx, db.GetThreadLastPosterAndPermsParams{
			ViewerID:      cd.UserID,
			ThreadID:      i,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.cache.forumThreadRows, &cd.cache.mapMu, id, fetch, ops...)
}

// ForumThread is a convenience wrapper around ForumThreadByID.
func (cd *CoreData) ForumThread(id int32, ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	return cd.ForumThreadByID(id, ops...)
}

// ForumThreads loads the threads for a forum topic once per topic.
func (cd *CoreData) ForumThreads(topicID int32) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	if cd.cache.forumThreads == nil {
		cd.cache.forumThreads = make(map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow])
	}
	lv, ok := cd.cache.forumThreads[topicID]
	if !ok {
		lv = &lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]{}
		cd.cache.forumThreads[topicID] = lv
	}
	return lv.Load(func() ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(cd.ctx, db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
			ViewerID:      cd.UserID,
			TopicID:       topicID,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// ForumTopicByID loads a forum topic once per ID using caching.
func (cd *CoreData) ForumTopicByID(id int32, ops ...lazy.Option[*db.GetForumTopicByIdForUserRow]) (*db.GetForumTopicByIdForUserRow, error) {
	fetch := func(i int32) (*db.GetForumTopicByIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumTopicByIdForUser(cd.ctx, db.GetForumTopicByIdForUserParams{
			ViewerID:      cd.UserID,
			Idforumtopic:  i,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.cache.forumTopics, &cd.cache.mapMu, id, fetch, ops...)
}

// ForumTopics loads forum topics for a given category once per category.
func (cd *CoreData) ForumTopics(categoryID int32) ([]*db.GetForumTopicsForUserRow, error) {
	if cd.cache.forumTopicLists == nil {
		cd.cache.forumTopicLists = make(map[int32]*lazy.Value[[]*db.GetForumTopicsForUserRow])
	}
	lv, ok := cd.cache.forumTopicLists[categoryID]
	if !ok {
		lv = &lazy.Value[[]*db.GetForumTopicsForUserRow]{}
		cd.cache.forumTopicLists[categoryID] = lv
	}
	return lv.Load(func() ([]*db.GetForumTopicsForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		if categoryID == 0 {
			return cd.queries.GetForumTopicsForUser(cd.ctx, db.GetForumTopicsForUserParams{ViewerID: cd.UserID, ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}})
		}
		rows, err := cd.queries.GetAllForumTopicsByCategoryIdForUserWithLastPosterName(cd.ctx, db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams{
			ViewerID:      cd.UserID,
			CategoryID:    categoryID,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			return nil, err
		}
		result := make([]*db.GetForumTopicsForUserRow, len(rows))
		for i, r := range rows {
			result[i] = &db.GetForumTopicsForUserRow{
				Idforumtopic:                 r.Idforumtopic,
				Lastposter:                   r.Lastposter,
				ForumcategoryIdforumcategory: r.ForumcategoryIdforumcategory,
				LanguageID:                   r.LanguageID,
				Title:                        r.Title,
				Description:                  r.Description,
				Threads:                      r.Threads,
				Comments:                     r.Comments,
				Lastaddition:                 r.Lastaddition,
				Handler:                      r.Handler,
				Lastposterusername:           r.Lastposterusername,
			}
		}
		return result, nil
	})
}

// ForumThreadReplies returns comments for the given thread.
func (cd *CoreData) ForumThreadReplies(threadID int32) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	return cd.ThreadComments(threadID)
}

// ForumComment loads a single comment by ID for the current user.
func (cd *CoreData) ForumComment(id int32) (*db.GetCommentByIdForUserRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	return cd.queries.GetCommentByIdForUser(cd.ctx, db.GetCommentByIdForUserParams{
		ViewerID: cd.UserID,
		ID:       id,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
}

// UpdateForumComment updates an existing comment owned by the current user.
func (cd *CoreData) UpdateForumComment(commentID, languageID int32, text string) error {
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
	if err := cd.validateImagePathsForThread(cd.UserID, comment.ForumthreadID, paths); err != nil {
		return fmt.Errorf("validate images: %w", err)
	}
	if err := cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   commentID,
		CommenterID: cd.UserID,
		EditorID:    sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	}); err != nil {
		return err
	}
	if err := cd.recordThreadImages(comment.ForumthreadID, paths); err != nil {
		log.Printf("record thread images: %v", err)
	}
	return nil
}

// EditForumComment updates a comment providing the commenter identifier explicitly.
func (cd *CoreData) EditForumComment(commentID, commenterID, languageID int32, text string) error {
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
		EditorID:    sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	}); err != nil {
		return err
	}
	if err := cd.recordThreadImages(comment.ForumthreadID, paths); err != nil {
		log.Printf("record thread images: %v", err)
	}
	return nil
}

func topicSubscriptionPattern(topicID int32) string {
	return fmt.Sprintf("%s:/forum/topic/%d/*", strings.ToLower("Create Thread"), topicID)
}

func threadSubscriptionPattern(topicID int32, threadID int32) string {
	return fmt.Sprintf("%s:/forum/topic/%d/thread/%d/*", strings.ToLower("Write Reply"), topicID, threadID)
}

// SubscribeTopic subscribes the current user to new threads in the given topic.
func (cd *CoreData) SubscribeTopic(userID, topicID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.InsertSubscription(cd.ctx, db.InsertSubscriptionParams{UsersIdusers: userID, Pattern: topicSubscriptionPattern(topicID), Method: "internal"})
}

// UnsubscribeTopic removes the current user's subscription to a topic.
func (cd *CoreData) UnsubscribeTopic(userID, topicID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.DeleteSubscriptionForSubscriber(cd.ctx, db.DeleteSubscriptionForSubscriberParams{SubscriberID: userID, Pattern: topicSubscriptionPattern(topicID), Method: "internal"})
}

// SubscribeThread subscribes the current user to new threads in the given thread.
func (cd *CoreData) SubscribeThread(topicID int32, threadID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.InsertSubscription(cd.ctx, db.InsertSubscriptionParams{UsersIdusers: cd.UserID, Pattern: threadSubscriptionPattern(topicID, threadID), Method: "internal"})
}

// UnsubscribeThread removes the current user's subscription to a thread.
func (cd *CoreData) UnsubscribeThread(topicID int32, threadID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.DeleteSubscriptionForSubscriber(cd.ctx, db.DeleteSubscriptionForSubscriberParams{SubscriberID: cd.UserID, Pattern: threadSubscriptionPattern(topicID, threadID), Method: "internal"})
}

// GrantForumCategory creates a grant for a forum category.
func (cd *CoreData) GrantForumCategory(categoryID int32, uid, rid sql.NullInt32, action string) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	if action == "" {
		action = "see"
	}
	return cd.queries.AdminCreateGrant(cd.ctx, db.AdminCreateGrantParams{
		UserID:   uid,
		RoleID:   rid,
		Section:  "forum",
		Item:     sql.NullString{String: "category", Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: categoryID, Valid: true},
		ItemRule: sql.NullString{},
		Action:   action,
		Extra:    sql.NullString{},
	})
}

// RevokeForumCategory removes a forum category grant by ID.
func (cd *CoreData) RevokeForumCategory(grantID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteGrant(cd.ctx, grantID)
}

// GrantForumTopic creates a grant for a forum topic.
func (cd *CoreData) GrantForumTopic(topicID int32, uid, rid sql.NullInt32, action string) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	if action == "" {
		action = "see"
	}
	return cd.queries.AdminCreateGrant(cd.ctx, db.AdminCreateGrantParams{
		UserID:   uid,
		RoleID:   rid,
		Section:  "forum",
		Item:     sql.NullString{String: "topic", Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
		ItemRule: sql.NullString{},
		Action:   action,
		Extra:    sql.NullString{},
	})
}

// RevokeForumTopic removes a forum topic grant by ID.
func (cd *CoreData) RevokeForumTopic(grantID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteGrant(cd.ctx, grantID)
}

// GrantForumThread creates a grant for a forum thread.
func (cd *CoreData) GrantForumThread(threadID int32, uid, rid sql.NullInt32, action string) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	if action == "" {
		action = "view"
	}
	return cd.queries.AdminCreateGrant(cd.ctx, db.AdminCreateGrantParams{
		UserID:   uid,
		RoleID:   rid,
		Section:  "privateforum",
		Item:     sql.NullString{String: "thread", Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: threadID, Valid: true},
		ItemRule: sql.NullString{},
		Action:   action,
		Extra:    sql.NullString{},
	})
}

// RevokeForumThread removes a forum thread grant by ID.
func (cd *CoreData) RevokeForumThread(grantID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteGrant(cd.ctx, grantID)
}
