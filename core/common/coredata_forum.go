package common

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
)

// ForumCategories loads all forum categories once.
func (cd *CoreData) ForumCategories() ([]*db.Forumcategory, error) {
	return cd.forumCategories.Load(func() ([]*db.Forumcategory, error) {
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
	return lazy.Map(&cd.forumThreadRows, &cd.mapMu, id, fetch, ops...)
}

// ForumThread is a convenience wrapper around ForumThreadByID.
func (cd *CoreData) ForumThread(id int32, ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	return cd.ForumThreadByID(id, ops...)
}

// ForumThreads loads the threads for a forum topic once per topic.
func (cd *CoreData) ForumThreads(topicID int32) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	if cd.forumThreads == nil {
		cd.forumThreads = make(map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow])
	}
	lv, ok := cd.forumThreads[topicID]
	if !ok {
		lv = &lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]{}
		cd.forumThreads[topicID] = lv
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
	return lazy.Map(&cd.forumTopics, &cd.mapMu, id, fetch, ops...)
}

// ForumTopics loads forum topics for a given category once per category.
func (cd *CoreData) ForumTopics(categoryID int32) ([]*db.GetForumTopicsForUserRow, error) {
	if cd.forumTopicLists == nil {
		cd.forumTopicLists = make(map[int32]*lazy.Value[[]*db.GetForumTopicsForUserRow])
	}
	lv, ok := cd.forumTopicLists[categoryID]
	if !ok {
		lv = &lazy.Value[[]*db.GetForumTopicsForUserRow]{}
		cd.forumTopicLists[categoryID] = lv
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
				LanguageIdlanguage:           r.LanguageIdlanguage,
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
	return cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   commentID,
		CommenterID: cd.UserID,
		EditorID:    sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
}

// EditForumComment updates a comment providing the commenter identifier explicitly.
func (cd *CoreData) EditForumComment(commentID, commenterID, languageID int32, text string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   commentID,
		CommenterID: commenterID,
		EditorID:    sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
}

func topicSubscriptionPattern(topicID int32) string {
	return fmt.Sprintf("%s:/forum/topic/%d/*", strings.ToLower("Create Thread"), topicID)
}

// SubscribeTopic subscribes the current user to new threads in the given topic.
func (cd *CoreData) SubscribeTopic(topicID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.InsertSubscription(cd.ctx, db.InsertSubscriptionParams{UsersIdusers: cd.UserID, Pattern: topicSubscriptionPattern(topicID), Method: "internal"})
}

// UnsubscribeTopic removes the current user's subscription to a topic.
func (cd *CoreData) UnsubscribeTopic(topicID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.DeleteSubscriptionForSubscriber(cd.ctx, db.DeleteSubscriptionForSubscriberParams{SubscriberID: cd.UserID, Pattern: topicSubscriptionPattern(topicID), Method: "internal"})
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

// ThreadPublicLabels returns public and owner labels for a thread sorted alphabetically.
func (cd *CoreData) ThreadPublicLabels(threadID int32) (public, owner []string, err error) {
	if cd.queries == nil {
		return nil, nil, nil
	}
	rows, err := cd.queries.ListTopicPublicLabels(cd.ctx, threadID)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range rows {
		public = append(public, r.Label)
	}
	ownerRows, err := cd.queries.ListContentLabelStatus(cd.ctx, db.ListContentLabelStatusParams{
		Item:   "forumtopic",
		ItemID: threadID,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, r := range ownerRows {
		owner = append(owner, r.Label)
	}
	sort.Strings(public)
	sort.Strings(owner)
	return public, owner, nil
}

// AddThreadPublicLabel adds a public label to a thread.
func (cd *CoreData) AddThreadPublicLabel(threadID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AddTopicPublicLabel(cd.ctx, db.AddTopicPublicLabelParams{
		ForumtopicIdforumtopic: threadID,
		Label:                  label,
	})
}

// RemoveThreadPublicLabel removes a public label from a thread.
func (cd *CoreData) RemoveThreadPublicLabel(threadID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.RemoveTopicPublicLabel(cd.ctx, db.RemoveTopicPublicLabelParams{
		ForumtopicIdforumtopic: threadID,
		Label:                  label,
	})
}

// AddThreadAuthorLabel adds an owner-only label to a thread.
func (cd *CoreData) AddThreadAuthorLabel(threadID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AddContentLabelStatus(cd.ctx, db.AddContentLabelStatusParams{
		Item:   "forumtopic",
		ItemID: threadID,
		Label:  label,
	})
}

// RemoveThreadAuthorLabel removes an owner-only label from a thread.
func (cd *CoreData) RemoveThreadAuthorLabel(threadID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.RemoveContentLabelStatus(cd.ctx, db.RemoveContentLabelStatusParams{
		Item:   "forumtopic",
		ItemID: threadID,
		Label:  label,
	})
}

// SetThreadPublicLabels replaces all public labels on a thread with the provided list.
func (cd *CoreData) SetThreadPublicLabels(threadID int32, labels []string) error {
	if cd.queries == nil {
		return nil
	}
	current, _, err := cd.ThreadPublicLabels(threadID)
	if err != nil {
		return err
	}
	want := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		want[l] = struct{}{}
	}
	have := make(map[string]struct{}, len(current))
	for _, l := range current {
		have[l] = struct{}{}
	}
	for l := range want {
		if _, ok := have[l]; !ok {
			if err := cd.AddThreadPublicLabel(threadID, l); err != nil {
				return err
			}
		}
	}
	for l := range have {
		if _, ok := want[l]; !ok {
			if err := cd.RemoveThreadPublicLabel(threadID, l); err != nil {
				return err
			}
		}
	}
	return nil
}

// ThreadPrivateLabels returns private labels for a thread sorted alphabetically.
// The special "new" and "unread" labels are stored separately and prefixed to the result.
func (cd *CoreData) ThreadPrivateLabels(threadID int32) ([]string, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.ListTopicPrivateLabels(cd.ctx, db.ListTopicPrivateLabelsParams{
		ForumtopicIdforumtopic: threadID,
		UsersIdusers:           cd.UserID,
	})
	if err != nil {
		return nil, err
	}
	var userLabels []string
	hasNew, hasUnread := true, true
	for _, r := range rows {
		switch r.Label {
		case "new":
			if r.Invert {
				hasNew = false
			} else {
				hasNew = true
			}
		case "unread":
			if r.Invert {
				hasUnread = false
			} else {
				hasUnread = true
			}
		default:
			if !r.Invert {
				userLabels = append(userLabels, r.Label)
			}
		}
	}
	sort.Strings(userLabels)
	labels := make([]string, 0, len(userLabels)+2)
	if hasNew {
		labels = append(labels, "new")
	}
	if hasUnread {
		labels = append(labels, "unread")
	}
	labels = append(labels, userLabels...)
	return labels, nil
}

// ClearThreadPrivateLabelStatus removes stored new/unread inversions for a thread across all users.
func (cd *CoreData) ClearThreadPrivateLabelStatus(threadID int32) error {
	if cd.queries == nil {
		return nil
	}
	if err := cd.queries.SystemClearTopicPrivateLabel(cd.ctx, db.SystemClearTopicPrivateLabelParams{
		ForumtopicIdforumtopic: threadID,
		Label:                  "new",
	}); err != nil {
		return err
	}
	return cd.queries.SystemClearTopicPrivateLabel(cd.ctx, db.SystemClearTopicPrivateLabelParams{
		ForumtopicIdforumtopic: threadID,
		Label:                  "unread",
	})
}

// SetThreadPrivateLabelStatus updates the special new/unread flags for a thread.
func (cd *CoreData) SetThreadPrivateLabelStatus(threadID int32, newLabel, unreadLabel bool) error {
	if cd.queries == nil {
		return nil
	}
	if newLabel {
		if err := cd.queries.RemoveTopicPrivateLabel(cd.ctx, db.RemoveTopicPrivateLabelParams{
			ForumtopicIdforumtopic: threadID,
			UsersIdusers:           cd.UserID,
			Label:                  "new",
		}); err != nil {
			return err
		}
	} else {
		if err := cd.queries.AddTopicPrivateLabel(cd.ctx, db.AddTopicPrivateLabelParams{
			ForumtopicIdforumtopic: threadID,
			UsersIdusers:           cd.UserID,
			Label:                  "new",
			Invert:                 true,
		}); err != nil {
			return err
		}
	}
	if unreadLabel {
		if err := cd.queries.RemoveTopicPrivateLabel(cd.ctx, db.RemoveTopicPrivateLabelParams{
			ForumtopicIdforumtopic: threadID,
			UsersIdusers:           cd.UserID,
			Label:                  "unread",
		}); err != nil {
			return err
		}
	} else {
		if err := cd.queries.AddTopicPrivateLabel(cd.ctx, db.AddTopicPrivateLabelParams{
			ForumtopicIdforumtopic: threadID,
			UsersIdusers:           cd.UserID,
			Label:                  "unread",
			Invert:                 true,
		}); err != nil {
			return err
		}
	}
	return nil
}

// AddThreadPrivateLabel adds a private label to a thread for the current user.
func (cd *CoreData) AddThreadPrivateLabel(threadID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AddTopicPrivateLabel(cd.ctx, db.AddTopicPrivateLabelParams{
		ForumtopicIdforumtopic: threadID,
		UsersIdusers:           cd.UserID,
		Label:                  label,
		Invert:                 false,
	})
}

// RemoveThreadPrivateLabel removes a private label from a thread for the current user.
func (cd *CoreData) RemoveThreadPrivateLabel(threadID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.RemoveTopicPrivateLabel(cd.ctx, db.RemoveTopicPrivateLabelParams{
		ForumtopicIdforumtopic: threadID,
		UsersIdusers:           cd.UserID,
		Label:                  label,
	})
}

// SetThreadPrivateLabels replaces all private labels for the current user on a thread with the provided list.
func (cd *CoreData) SetThreadPrivateLabels(threadID int32, labels []string) error {
	if cd.queries == nil {
		return nil
	}
	rows, err := cd.queries.ListTopicPrivateLabels(cd.ctx, db.ListTopicPrivateLabelsParams{
		ForumtopicIdforumtopic: threadID,
		UsersIdusers:           cd.UserID,
	})
	if err != nil {
		return err
	}
	have := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if r.Invert || r.Label == "new" || r.Label == "unread" {
			continue
		}
		have[r.Label] = struct{}{}
	}
	want := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		want[l] = struct{}{}
	}
	for l := range want {
		if _, ok := have[l]; !ok {
			if err := cd.AddThreadPrivateLabel(threadID, l); err != nil {
				return err
			}
		}
	}
	for l := range have {
		if _, ok := want[l]; !ok {
			if err := cd.RemoveThreadPrivateLabel(threadID, l); err != nil {
				return err
			}
		}
	}
	return nil
}
