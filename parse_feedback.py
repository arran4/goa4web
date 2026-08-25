# 1. Atomic append in SQL.
# MySQL: SET text = CONCAT(c.text, '\n\n[hr]\n\n', sqlc.arg(text))
# SQLite: SET text = comments.text || '\n\n[hr]\n\n' || sqlc.arg(text)
# We need to change queries-comments.sql and dbsqlite_queries/queries-comments.sql
#
# 2. Preserve normal image pipeline: we already sanitize, queue, validate, and record images. Wait, `AttemptAppendForumComment` currently doesn't call `recordThreadImages` or `StartRemoteImageCacheFetch`!
# Let's fix that.
#
# 3. Refresh thread/topic activity without increasing post count.
# I had disabled `IncludePostCount` in `HandleThreadUpdated`. But the feedback says: "append suppresses IncludePostCount, but the post-count worker also recalculates thread metadata and topic aggregates. Do not leave forumthread.lastaddition or topic metadata stale... Recalculation is fine because the number of actual comment rows has not changed, so the comment count should remain unchanged naturally."
# So I should SET `IncludePostCount: true`.
#
# 4. Public `append` grant.
# Wait, I did add it to "forum|topic": {Actions: []string{"see", "view", "reply", "post", "edit", "append"} in `handlers/admin/role_grants.go`. I should double check it's definitely there.
#
# 5. Fix `RequireCommentAuthor` topic ID check.
# `cd.HasGrant(section, "topic", "edit", cd.SelectedThreadID())` -> this is wrong! The item ID for topic is the TOPIC ID, not the THREAD ID.
# So I need to use `cd.CurrentTopicID()` or something instead of `cd.SelectedThreadID()` for the topic grant checks!
#
# 6. Page-time `CanAppendToComment` read-marker check.
# I left a comment saying "we'll omit a direct read marker query here". I need to actually query it.
# I can use `cd.queries.GetContentReadMarker`? Wait, I need to know if ANY OTHER user has a read marker >= comment ID.
# There is a query `GetReadMarkersForThreadAndComment` or I can write a helper query, or just loop through all read markers?
# Wait, there's no query for "all read markers > X". I might need to add one. Or look at `queries-read_markers.sql` and add `HasOtherUserReadItem` query.
#
# 7. Search indexing.
# The search index gets `text`. Since the SQL now CONCATs the text, the `text` we pass to `HandleThreadUpdated` is only the NEW segment.
# If we pass only the new segment, the search worker will index only the new segment and clear the old words!
# We MUST pass the full text. So we must do `newFullText = existingText + "\n\n[hr]\n\n" + newText` in Go and pass `newFullText` to `HandleThreadUpdated` (CommentText field).
#
# 8. Regression suite.
# Let's write the massive test suite!
