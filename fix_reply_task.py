import sys

with open("handlers/forum/forumTopicThreadReplyPage.go", "r") as f:
    content = f.read()

old_reply = """	// Check if this might be an append. We need the last comment ID.
	commentsList, _ := cd.ThreadComments(threadRow.Idforumthread)
	if len(commentsList) > 0 {
	    lastComment := commentsList[len(commentsList)-1]
	    // Attempt append first
	    cid, err = cd.AttemptAppendForumComment(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), lastComment.Idcomments, text, topicRow.Handler == "private")
	    if err != nil {
	        log.Printf("Append attempt error: %v", err)
	    }
	    if cid != 0 {
	        isAppend = true
	    }
	}"""

new_reply = """	// Check if this might be an append. We need the last comment ID.
	var fullText string
	commentsList, _ := cd.ThreadComments(threadRow.Idforumthread)
	if len(commentsList) > 0 {
	    lastComment := commentsList[len(commentsList)-1]
	    // Attempt append first
	    cid, fullText, err = cd.AttemptAppendForumComment(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), lastComment.Idcomments, text, topicRow.Handler == "private")
	    if err != nil {
	        log.Printf("Append attempt error: %v", err)
	    }
	    if cid != 0 {
	        isAppend = true
	        text = fullText // update the text for the thread updated event
	    }
	}"""

content = content.replace(old_reply, new_reply)

# Update HandleThreadUpdated to IncludePostCount: true
old_handle = """		IncludePostCount:     !isAppend,
		IncludeSearch:        true,"""
new_handle = """		IncludePostCount:     true,
		IncludeSearch:        true,"""

content = content.replace(old_handle, new_handle)

with open("handlers/forum/forumTopicThreadReplyPage.go", "w") as f:
    f.write(content)
