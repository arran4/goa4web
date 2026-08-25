import sys

with open("core/common/coredata.go", "r") as f:
    content = f.read()

# 1. Update AttemptAppendForumComment to only pass the new text to SQL (the CONCAT happens in SQL).
# Wait, search worker needs the FULL text.
# So we need to return the FULL text from `AttemptAppendForumComment` to be used by the event.
# And we need to remove the `[hr]` concatenation before SQL update since SQL does it now! Wait, `text` in `sqlc.arg(text)` is the NEW segment, but SQL does `CONCAT(c.text, '[hr]', text)`.
# But `sanitizeCodeImages` is currently sanitizing only the new segment (which is correct!).
# However, if we need the full text for search, `AttemptAppendForumComment` could just return it.
# Let's change `AttemptAppendForumComment` signature to return `(int64, string, error)`.

old_attempt = """func (cd *CoreData) AttemptAppendForumComment(commenterID int32, threadID int32, topicID int32, languageID int32, commentID int32, text string, isPrivate bool) (int64, error) {"""
new_attempt = """func (cd *CoreData) AttemptAppendForumComment(commenterID int32, threadID int32, topicID int32, languageID int32, commentID int32, text string, isPrivate bool) (int64, string, error) {"""
content = content.replace(old_attempt, new_attempt)

old_attempt_body = """	if cd.queries == nil || cd.Config == nil {
		return 0, nil
	}
	appendWindowMins := cd.Config.ForumPostAppendWindow
	section := consts.PermissionSectionForum.String()
	itemType := consts.PermissionItemTopic.String()
	itemID := topicID
	if isPrivate {
	    appendWindowMins = cd.Config.PrivateForumPostAppendWindow
	    section = consts.PermissionSectionPrivateForumThread.String()
	    itemType = consts.PermissionItemThread.String()
	    itemID = threadID
	}
	if appendWindowMins <= 0 {
	    return 0, nil
	}
	text = cd.sanitizeCodeImages(text)
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return 0, fmt.Errorf("parse images: %w", err)
	}
	comment, err := cd.CommentByID(commentID)
	if err != nil || comment == nil {
	    return 0, fmt.Errorf("load comment: %w", err)
	}
	if err := cd.validateImagePathsForThread(cd.UserID, comment.ForumthreadID, paths); err != nil {
		return 0, fmt.Errorf("validate images: %w", err)
	}
	newText := fmt.Sprintf("%s\\n\\n[hr]\\n\\n%s", comment.Text.String, text)
	rowsAffected, err := cd.queries.AppendCommentInSectionForCommenter(cd.ctx, db.AppendCommentInSectionForCommenterParams{
		Text:             sql.NullString{String: newText, Valid: true},
		Written:          sql.NullTime{Time: time.Now(), Valid: true},
		CommentID:        commentID,
		CommenterID:      commenterID,
		ForumthreadID:    threadID,
		AppendWindowMins: int64(appendWindowMins),
		Section:          section,
		ItemType:         sql.NullString{String: itemType, Valid: true},
		ItemID:           sql.NullInt32{Int32: itemID, Valid: true},
		GrantUserID:      sql.NullInt32{Int32: commenterID, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	if rowsAffected > 0 {
	    return int64(commentID), nil
	}
	return 0, nil"""

new_attempt_body = """	if cd.queries == nil || cd.Config == nil {
		return 0, "", nil
	}
	appendWindowMins := cd.Config.ForumPostAppendWindow
	section := consts.PermissionSectionForum.String()
	itemType := consts.PermissionItemTopic.String()
	itemID := topicID
	if isPrivate {
	    appendWindowMins = cd.Config.PrivateForumPostAppendWindow
	    section = consts.PermissionSectionPrivateForumThread.String()
	    itemType = consts.PermissionItemThread.String()
	    itemID = threadID
	}
	if appendWindowMins <= 0 {
	    return 0, "", nil
	}

	// Preserve the normal image pipeline
	var queuedFetches []queuedRemoteImageCacheFetch
	text, queuedFetches = cd.sanitizeCodeImagesAndQueue(text)
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return 0, "", fmt.Errorf("parse images: %w", err)
	}
	comment, err := cd.CommentByID(commentID)
	if err != nil || comment == nil {
	    return 0, "", fmt.Errorf("load comment: %w", err)
	}
	if err := cd.validateImagePathsForThread(cd.UserID, comment.ForumthreadID, paths); err != nil {
		return 0, "", fmt.Errorf("validate images: %w", err)
	}

	rowsAffected, err := cd.queries.AppendCommentInSectionForCommenter(cd.ctx, db.AppendCommentInSectionForCommenterParams{
		Text:             sql.NullString{String: text, Valid: true},
		Written:          sql.NullTime{Time: time.Now(), Valid: true},
		CommentID:        commentID,
		CommenterID:      commenterID,
		ForumthreadID:    threadID,
		AppendWindowMins: int64(appendWindowMins),
		Section:          section,
		ItemType:         sql.NullString{String: itemType, Valid: true},
		ItemID:           sql.NullInt32{Int32: itemID, Valid: true},
		GrantUserID:      sql.NullInt32{Int32: commenterID, Valid: true},
	})
	if err != nil {
		return 0, "", err
	}
	if rowsAffected > 0 {
	    // Queue required remote-image cache fetches
	    for _, fetch := range queuedFetches {
		    cd.StartRemoteImageCacheFetch(fetch.id, fetch.sourceURL)
	    }
	    // Record thread-image associations
	    if err := cd.recordThreadImages(threadID, paths); err != nil {
		    log.Printf("record thread images append: %v", err)
	    }
	    // Reconstruct full text for search indexing
	    fullText := fmt.Sprintf("%s\\n\\n[hr]\\n\\n%s", comment.Text.String, text)
	    return int64(commentID), fullText, nil
	}
	return 0, "", nil"""

content = content.replace(old_attempt_body, new_attempt_body)

# 2. Update CanAppendToComment to check HasOtherUserReadItemAtOrBeyond
old_read_marker = """	if cd.queries != nil {
	    // Technically we shouldn't execute direct queries from UI helpers as much, but we need
	    // to check content read markers. We could loop manually or trust the attempt fails on POST.
	    // However, the page display needs this to show "Append" reliably.
	    // We'll trust the POST attempt for absolute safety, but we can do a quick check here.

	    // Note: Due to limitations in the current stub/query structure, we'll omit a direct read marker query here
	    // to prevent over-fetching on every comment render, or simply return true if other conditions met,
	    // trusting the backend to downgrade to a reply if a read marker blocks it.
	}"""

new_read_marker = """	if cd.queries != nil {
	    hasRead, _ := cd.queries.HasOtherUserReadItemAtOrBeyond(cd.ctx, db.HasOtherUserReadItemAtOrBeyondParams{
	        Item: "thread",
	        ItemID: cmt.ForumthreadID,
	        UserID: cd.UserID,
	        LastCommentID: cmt.Idcomments,
	    })
	    if hasRead {
	        return false
	    }
	}"""

content = content.replace(old_read_marker, new_read_marker)


with open("core/common/coredata.go", "w") as f:
    f.write(content)
