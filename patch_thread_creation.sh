#!/bin/bash

# Modify handlers/forum/forumThreadNewPage.go to process quote_comment_id during thread creation
sed -i '/text := r.PostFormValue("replytext")/i \
	quoteCommentId := r.URL.Query().Get("quote_comment_id")\
	var replyToCommentId, replyToThreadId sql.NullInt32\
	if quoteCommentId != "" {\
		if cId, err := strconv.Atoi(quoteCommentId); err == nil {\
			if c, err := cd.CommentByID(int32(cId)); err == nil \&\& c != nil {\
				replyToCommentId = sql.NullInt32{Int32: int32(cId), Valid: true}\
				replyToThreadId = sql.NullInt32{Int32: c.ForumthreadID, Valid: true}\
			}\
		}\
	}\
' handlers/forum/forumThreadNewPage.go

sed -i '/if evt := cd.Event(); evt != nil {/i \
	if replyToThreadId.Valid {\
		if err := cd.Queries().SetThreadReplyTo(r.Context(), db.SetThreadReplyToParams{\
			ReplyToCommentId: replyToCommentId,\
			ReplyToThreadId:  replyToThreadId,\
			Idforumthread:    int32(threadId),\
		}); err != nil {\
			log.Printf("Error: setting thread reply to: %s", err)\
		}\
	}\
' handlers/forum/forumThreadNewPage.go
