import sys

with open("handlers/forum/forumTopicThreadReplyPage.go", "r") as f:
    content = f.read()

bad = """	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             threadRow.Idforumthread,
		TopicID:              topicRow.Idforumtopic,
		CommentID:            int32(cid),
		Thread:               threadRow,
		TopicTitle:           topicRow.Title.String,
		Username:             username,
		CommentText:          text,
		CommentURL:           cd.AbsoluteURL(endUrl),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
		AdditionalData:       data,
	}); err != nil {"""

good = """	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             threadRow.Idforumthread,
		TopicID:              topicRow.Idforumtopic,
		CommentID:            int32(cid),
		Thread:               threadRow,
		TopicTitle:           topicRow.Title.String,
		Username:             username,
		CommentText:          text,
		CommentURL:           cd.AbsoluteURL(endUrl),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     !isAppend,
		IncludeSearch:        true,
		AdditionalData:       data,
	}); err != nil {"""

content = content.replace(bad, good)
with open("handlers/forum/forumTopicThreadReplyPage.go", "w") as f:
    f.write(content)
