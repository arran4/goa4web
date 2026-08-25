import sys

with open("core/common/coredata.go", "r") as f:
    content = f.read()

bad_sig = "func (cd *CoreData) AttemptAppendForumComment(commenterID, threadID, topicID, languageID, commentID int32, text string, isPrivate bool) (int64, error) {"
good_sig = "func (cd *CoreData) AttemptAppendForumComment(commenterID int32, threadID int32, topicID int32, languageID int32, commentID int32, text string, isPrivate bool) (int64, string, error) {"

content = content.replace(bad_sig, good_sig)
with open("core/common/coredata.go", "w") as f:
    f.write(content)
