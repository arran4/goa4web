package common

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// ForumOperationForbiddenError reports that an actor lacks permission for a forum operation.
type ForumOperationForbiddenError struct {
	Action string
}

func (e ForumOperationForbiddenError) Error() string {
	if e.Action == "" {
		return "forum operation forbidden"
	}
	return fmt.Sprintf("forum %s forbidden", e.Action)
}

// ForumResourceNotFoundError reports that a requested forum resource is not visible to the actor.
type ForumResourceNotFoundError struct {
	Resource string
}

func (e ForumResourceNotFoundError) Error() string {
	if e.Resource == "" {
		return "forum resource not found"
	}
	return fmt.Sprintf("forum %s not found", e.Resource)
}

// ForumHandlerMismatchError reports that public/private route semantics do not match a topic.
type ForumHandlerMismatchError struct {
	ExpectedPrivate bool
}

func (e ForumHandlerMismatchError) Error() string {
	if e.ExpectedPrivate {
		return "topic is not a private forum topic"
	}
	return "private forum topic cannot be used through the public forum"
}

// CreateForumThreadParams describes a normal forum or private-forum thread creation.
type CreateForumThreadParams struct {
	ActorID                int32
	TopicID                int32
	LanguageID             int32
	Text                   string
	At                     time.Time
	Private                bool
	EnforceHandler         bool
	SynchronousSideEffects bool
	BasePath               string
	PublicLabels           []string
	PrivateLabels          []string
	ReplyToCommentID       int32
	ReplyToThreadID        int32
}

// CreateForumThreadResult identifies the thread and opening comment created by CreateForumThread.
type CreateForumThreadResult struct {
	ThreadID  int32
	CommentID int32
	URL       string
}

// ReplyForumThreadParams describes a normal reply to a forum or private-forum thread.
type ReplyForumThreadParams struct {
	ActorID                int32
	ThreadID               int32
	LanguageID             int32
	Text                   string
	At                     time.Time
	Private                bool
	EnforceHandler         bool
	SynchronousSideEffects bool
	BasePath               string
}

// ReplyForumThreadResult identifies the comment created by ReplyForumThread.
type ReplyForumThreadResult struct {
	ThreadID  int32
	TopicID   int32
	CommentID int32
	URL       string
	Appended  bool
}

type forumCommentAppendResult struct {
	appended      bool
	canonicalText string
	textAvailable bool
}

// CanCreateForumThread reports whether actorID may post a thread in topicID.
func CanCreateForumThread(ctx context.Context, q db.Querier, section consts.PermissionSection, topicID, actorID int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID:               actorID,
		Section:                section.String(),
		Item:                   sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		Action:                 consts.PermissionActionPost.String(),
		ItemID:                 sql.NullInt32{Int32: topicID, Valid: true},
		IsSpecificPrivateForum: (section == consts.PermissionSectionPrivateForum || section == consts.PermissionSectionPrivateForumThread) && topicID != 0,
		UserID:                 sql.NullInt32{Int32: actorID, Valid: actorID != 0},
	})
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

// CanReplyForumThread reports whether actorID may reply to a visible, unlocked thread.
func CanReplyForumThread(ctx context.Context, q db.Querier, section consts.PermissionSection, itemType consts.PermissionItem, itemID, threadID, actorID int32) (bool, error) {
	thread, err := q.GetThreadBySectionThreadIDForReplier(ctx, db.GetThreadBySectionThreadIDForReplierParams{
		ReplierID:      actorID,
		ThreadID:       threadID,
		Section:        section.String(),
		ItemType:       sql.NullString{String: itemType.String(), Valid: true},
		ItemID:         sql.NullInt32{Int32: itemID, Valid: itemID != 0},
		ReplierMatchID: sql.NullInt32{Int32: actorID, Valid: actorID != 0},
	})
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return thread != nil && (!thread.Locked.Valid || !thread.Locked.Bool), nil
}

func forumBasePath(private bool, basePath string) string {
	if basePath != "" {
		return basePath
	}
	if private {
		return "/private"
	}
	return "/forum"
}

func (cd *CoreData) forumAppendWindow(isPrivate bool) int {
	if cd == nil || cd.Config == nil {
		return 0
	}
	if isPrivate {
		return cd.Config.PrivateForumPostAppendWindow
	}
	return cd.Config.ForumPostAppendWindow
}

func forumAppendGrantTarget(isPrivate bool, topicID, threadID int32) (consts.PermissionSection, consts.PermissionItem, int32) {
	if isPrivate {
		return consts.PermissionSectionPrivateForumThread, consts.PermissionItemThread, threadID
	}
	return consts.PermissionSectionForum, consts.PermissionItemTopic, topicID
}

func (cd *CoreData) attemptAppendForumComment(ctx context.Context, actorID, threadID, topicID, commentID int32, text string, writtenAt time.Time, isPrivate bool) (forumCommentAppendResult, error) {
	appendWindowMins := cd.forumAppendWindow(isPrivate)
	if appendWindowMins <= 0 {
		return forumCommentAppendResult{}, nil
	}

	text, queuedFetches := cd.sanitizeCodeImagesAndQueue(text)
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return forumCommentAppendResult{}, fmt.Errorf("parse images: %w", imageValidationUserError(err))
	}
	if err := cd.validateImagePathsForThread(actorID, threadID, paths); err != nil {
		return forumCommentAppendResult{}, fmt.Errorf("validate images: %w", err)
	}

	section, itemType, itemID := forumAppendGrantTarget(isPrivate, topicID, threadID)
	rowsAffected, err := cd.queries.AppendCommentInSectionForCommenter(ctx, db.AppendCommentInSectionForCommenterParams{
		Text:             text,
		Written:          sql.NullTime{Time: writtenAt.UTC(), Valid: true},
		CommentID:        commentID,
		CommenterID:      actorID,
		ForumthreadID:    threadID,
		AppendWindowMins: int64(appendWindowMins),
		Section:          section.String(),
		ItemType:         sql.NullString{String: itemType.String(), Valid: true},
		ItemID:           sql.NullInt32{Int32: itemID, Valid: true},
		GrantUserID:      sql.NullInt32{Int32: actorID, Valid: actorID != 0},
	})
	if err != nil {
		return forumCommentAppendResult{}, fmt.Errorf("append forum reply: %w", err)
	}
	if rowsAffected == 0 {
		return forumCommentAppendResult{}, nil
	}

	for _, fetch := range queuedFetches {
		cd.StartRemoteImageCacheFetch(fetch.id, fetch.sourceURL)
	}
	if err := cd.recordThreadImages(threadID, paths); err != nil {
		log.Printf("record thread images after append: %v", err)
	}

	comment, err := cd.queries.GetCommentByIdForUser(ctx, db.GetCommentByIdForUserParams{
		ViewerID: actorID,
		ID:       commentID,
		UserID:   sql.NullInt32{Int32: actorID, Valid: actorID != 0},
	})
	if err != nil || comment == nil {
		if err != nil {
			log.Printf("reload appended forum comment %d: %v", commentID, err)
		}
		return forumCommentAppendResult{appended: true}, nil
	}
	return forumCommentAppendResult{
		appended:      true,
		canonicalText: comment.Text.String,
		textAvailable: true,
	}, nil
}

func (cd *CoreData) forumTopicForActor(ctx context.Context, topicID, actorID int32) (*db.GetForumTopicByIdForUserRow, error) {
	topic, err := cd.queries.GetForumTopicByIdForUser(ctx, db.GetForumTopicByIdForUserParams{
		ViewerID:      actorID,
		Idforumtopic:  topicID,
		ViewerMatchID: sql.NullInt32{Int32: actorID, Valid: actorID != 0},
	})
	if err == sql.ErrNoRows || topic == nil {
		return nil, ForumResourceNotFoundError{Resource: "topic"}
	}
	if err != nil {
		return nil, fmt.Errorf("get forum topic for actor: %w", err)
	}
	return topic, nil
}

func (cd *CoreData) cleanupUninitializedForumThread(ctx context.Context, threadID int32, cause error) error {
	grants, grantsErr := cd.queries.AdminListGrantsByThreadID(ctx, sql.NullInt32{Int32: threadID, Valid: true})
	if grantsErr == nil {
		for _, grant := range grants {
			if err := cd.queries.AdminDeleteGrant(ctx, grant.ID); err != nil {
				grantsErr = fmt.Errorf("delete grant %d: %w", grant.ID, err)
				break
			}
		}
	}
	if err := cd.queries.SystemDeleteUninitializedThread(ctx, threadID); err != nil {
		return fmt.Errorf("%w; cleanup uninitialized thread %d: %v", cause, threadID, err)
	}
	if grantsErr != nil {
		return fmt.Errorf("%w; cleanup grants for uninitialized thread %d: %v", cause, threadID, grantsErr)
	}
	return cause
}

// CreateForumThread creates a thread and opening comment through the same permission and side-effect boundary used by HTTP handlers and scenario imports.
func (cd *CoreData) CreateForumThread(ctx context.Context, params CreateForumThreadParams) (*CreateForumThreadResult, error) {
	if cd == nil || cd.queries == nil {
		return nil, fmt.Errorf("create forum thread: no queries")
	}
	topic, err := cd.forumTopicForActor(ctx, params.TopicID, params.ActorID)
	if err != nil {
		return nil, err
	}
	isPrivate := topic.Handler == "private"
	if params.EnforceHandler && params.Private != isPrivate {
		return nil, ForumHandlerMismatchError{ExpectedPrivate: params.Private}
	}
	section := consts.PermissionSectionForum
	if isPrivate {
		section = consts.PermissionSectionPrivateForum
	}
	allowed, err := CanCreateForumThread(ctx, cd.queries, section, params.TopicID, params.ActorID)
	if err != nil {
		return nil, fmt.Errorf("check forum thread create permission: %w", err)
	}
	if !allowed {
		return nil, ForumOperationForbiddenError{Action: "thread creation"}
	}

	var threadID64 int64
	if params.ReplyToThreadID != 0 {
		threadID64, err = cd.queries.SystemCreateReplyThread(ctx, db.SystemCreateReplyThreadParams{
			TopicID:          params.TopicID,
			ReplyToCommentID: sql.NullInt32{Int32: params.ReplyToCommentID, Valid: params.ReplyToCommentID != 0},
			ReplyToThreadID:  sql.NullInt32{Int32: params.ReplyToThreadID, Valid: true},
		})
	} else {
		threadID64, err = cd.queries.SystemCreateThread(ctx, params.TopicID)
	}
	if err != nil {
		return nil, fmt.Errorf("create forum thread row: %w", err)
	}
	threadID := int32(threadID64)

	if isPrivate {
		if params.ReplyToThreadID != 0 {
			err = cd.CopyPrivateThreadGrantsToThread(params.ReplyToThreadID, threadID)
		} else {
			err = cd.CopyPrivateTopicGrantsToThread(params.TopicID, threadID)
		}
		if err != nil {
			return nil, cd.cleanupUninitializedForumThread(ctx, threadID, fmt.Errorf("copy private grants to thread: %w", err))
		}
	}

	writtenAt := params.At
	if writtenAt.IsZero() {
		writtenAt = time.Now().UTC()
	}
	commentSection := consts.PermissionSectionForum
	if isPrivate {
		commentSection = consts.PermissionSectionPrivateForum
	}
	commentID64, err := cd.createCommentInSectionForCommenterAt(
		commentSection, consts.PermissionItemTopic, consts.PermissionActionPost,
		params.TopicID, threadID, params.ActorID, params.LanguageID, params.Text, writtenAt,
	)
	if err != nil {
		return nil, cd.cleanupUninitializedForumThread(ctx, threadID, fmt.Errorf("create forum opening comment: %w", err))
	}
	if commentID64 == 0 {
		return nil, cd.cleanupUninitializedForumThread(ctx, threadID, ForumOperationForbiddenError{Action: "opening comment creation"})
	}
	commentID := int32(commentID64)

	if err := cd.SetThreadPublicLabels(threadID, params.PublicLabels); err != nil {
		log.Printf("set public labels: %v", err)
	}
	if err := cd.SetThreadPrivateLabels(threadID, params.PrivateLabels); err != nil {
		log.Printf("set private labels: %v", err)
	}

	basePath := forumBasePath(isPrivate, params.BasePath)
	endURL := fmt.Sprintf("%s/topic/%d/thread/%d", basePath, params.TopicID, threadID)
	username := ""
	if user := cd.UserByID(params.ActorID); user != nil {
		username = user.Username.String
	}
	subjectPrefix := "Forum"
	if isPrivate {
		subjectPrefix = "Private Forum"
	}
	if evt := cd.Event(); evt != nil {
		evt.Path = endURL
	}
	if err := cd.HandleThreadUpdated(ctx, ThreadUpdatedEvent{
		ThreadID:         threadID,
		TopicID:          params.TopicID,
		CommentID:        commentID,
		TopicTitle:       topic.Title.String,
		Author:           username,
		Username:         username,
		CommentText:      params.Text,
		PostURL:          cd.AbsoluteURL(endURL),
		ThreadURL:        cd.AbsoluteURL(endURL),
		IncludePostCount: true,
		IncludeSearch:    true,
		MarkThreadRead:   true,
		AdditionalData: map[string]any{
			"ThreadOpenerPreview": a4code.SnipTextWords(params.Text, 10),
			"SubjectPrefix":       subjectPrefix,
		},
	}); err != nil {
		log.Printf("thread create side effects: %v", err)
	}
	if params.SynchronousSideEffects {
		if err := cd.ApplyForumMutationWorkers(ctx, threadID, params.TopicID, commentID, params.Text); err != nil {
			return nil, fmt.Errorf("apply forum thread workers: %w", err)
		}
	}

	return &CreateForumThreadResult{ThreadID: threadID, CommentID: commentID, URL: endURL}, nil
}

// ReplyForumThread creates a reply through the same permission and side-effect boundary used by HTTP handlers and scenario imports.
func (cd *CoreData) ReplyForumThread(ctx context.Context, params ReplyForumThreadParams) (*ReplyForumThreadResult, error) {
	if cd == nil || cd.queries == nil {
		return nil, fmt.Errorf("reply to forum thread: no queries")
	}
	thread, err := cd.queries.GetThreadLastPosterAndPermsForUser(ctx, db.GetThreadLastPosterAndPermsForUserParams{
		ViewerID:      params.ActorID,
		ThreadID:      params.ThreadID,
		ViewerMatchID: sql.NullInt32{Int32: params.ActorID, Valid: params.ActorID != 0},
	})
	if err == sql.ErrNoRows || thread == nil {
		return nil, ForumResourceNotFoundError{Resource: "thread"}
	}
	if err != nil {
		return nil, fmt.Errorf("get forum thread for actor: %w", err)
	}
	topic, err := cd.forumTopicForActor(ctx, thread.ForumtopicIdforumtopic, params.ActorID)
	if err != nil {
		return nil, err
	}
	isPrivate := topic.Handler == "private"
	if params.EnforceHandler && params.Private != isPrivate {
		return nil, ForumHandlerMismatchError{ExpectedPrivate: params.Private}
	}
	section := consts.PermissionSectionForum
	itemType := consts.PermissionItemTopic
	itemID := topic.Idforumtopic
	if isPrivate {
		section = consts.PermissionSectionPrivateForumThread
		itemType = consts.PermissionItemThread
		itemID = thread.Idforumthread
	}
	allowed, err := CanReplyForumThread(ctx, cd.queries, section, itemType, itemID, thread.Idforumthread, params.ActorID)
	if err != nil {
		return nil, fmt.Errorf("check forum reply permission: %w", err)
	}
	if !allowed {
		return nil, ForumOperationForbiddenError{Action: "reply"}
	}

	writtenAt := params.At
	if writtenAt.IsZero() {
		writtenAt = time.Now().UTC()
	}

	var comments []*db.GetCommentsByThreadIdForUserRow
	commentsLoaded := false
	appendResult := forumCommentAppendResult{}
	if cd.forumAppendWindow(isPrivate) > 0 {
		comments, err = cd.ThreadComments(thread.Idforumthread)
		if err != nil {
			return nil, fmt.Errorf("load forum replies for append candidate: %w", err)
		}
		commentsLoaded = true
		if len(comments) > 0 {
			candidate := comments[len(comments)-1]
			appendResult, err = cd.attemptAppendForumComment(
				ctx,
				params.ActorID,
				thread.Idforumthread,
				topic.Idforumtopic,
				candidate.Idcomments,
				params.Text,
				writtenAt,
				isPrivate,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	commentID := int32(0)
	commentText := params.Text
	canonicalTextAvailable := true
	if appendResult.appended {
		commentID = comments[len(comments)-1].Idcomments
		commentText = appendResult.canonicalText
		canonicalTextAvailable = appendResult.textAvailable
	} else {
		commentID64, createErr := cd.createCommentInSectionForCommenterAt(
			section, itemType, consts.PermissionActionReply, itemID,
			thread.Idforumthread, params.ActorID, params.LanguageID, params.Text, writtenAt,
		)
		if createErr != nil {
			return nil, fmt.Errorf("create forum reply: %w", createErr)
		}
		if commentID64 == 0 {
			return nil, ForumOperationForbiddenError{Action: "reply"}
		}
		commentID = int32(commentID64)
	}

	anchor := fmt.Sprintf("c%d", commentID)
	if commentsLoaded && len(comments) > 0 {
		anchorIndex := len(comments)
		if !appendResult.appended {
			anchorIndex++
		}
		if anchorIndex > 0 {
			anchor = fmt.Sprintf("c%d", anchorIndex)
		}
	} else if currentComments, commentsErr := cd.ThreadComments(thread.Idforumthread); commentsErr != nil {
		log.Printf("fetch comments to determine reply anchor: %v", commentsErr)
	} else if len(currentComments) > 0 {
		anchor = fmt.Sprintf("c%d", len(currentComments))
	}
	basePath := forumBasePath(isPrivate, params.BasePath)
	endURL := fmt.Sprintf("%s/topic/%d/thread/%d#%s", basePath, topic.Idforumtopic, thread.Idforumthread, anchor)
	username := ""
	if user := cd.UserByID(params.ActorID); user != nil {
		username = user.Username.String
	}
	data := map[string]any{}
	if firstPost, firstPostErr := cd.CommentByID(thread.Firstpost); firstPostErr == nil && firstPost != nil && firstPost.Text.Valid {
		data["ThreadOpenerPreview"] = a4code.SnipTextWords(firstPost.Text.String, 10)
	}
	if isPrivate {
		data["SubjectPrefix"] = "Private Forum"
	} else {
		data["SubjectPrefix"] = "Forum"
	}
	if err := cd.HandleThreadUpdated(ctx, ThreadUpdatedEvent{
		ThreadID:             thread.Idforumthread,
		TopicID:              topic.Idforumtopic,
		CommentID:            commentID,
		Thread:               thread,
		TopicTitle:           topic.Title.String,
		Username:             username,
		CommentText:          commentText,
		CommentURL:           cd.AbsoluteURL(endURL),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        canonicalTextAvailable,
		AdditionalData:       data,
	}); err != nil {
		log.Printf("thread reply side effects: %v", err)
	}
	if params.SynchronousSideEffects {
		if err := cd.applyForumMutationWorkers(ctx, thread.Idforumthread, topic.Idforumtopic, commentID, commentText, canonicalTextAvailable); err != nil {
			return nil, fmt.Errorf("apply forum reply workers: %w", err)
		}
	}

	return &ReplyForumThreadResult{
		ThreadID:  thread.Idforumthread,
		TopicID:   topic.Idforumtopic,
		CommentID: commentID,
		URL:       endURL,
		Appended:  appendResult.appended,
	}, nil
}
