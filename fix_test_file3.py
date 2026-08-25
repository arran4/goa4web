import sys

with open("handlers/forum/forum_post_append_test.go", "r") as f:
    content = f.read()

content = content.replace("q.GetCommentByIdFn =", "// q.GetCommentByIdFn =")
content = content.replace("q.AppendCommentInSectionForCommenterCalls", "q.AppendCommentInSectionForCommenterCalls") # Actually they might not be added to QuerierStub because they are manual

with open("handlers/forum/forum_post_append_test.go", "w") as f:
    f.write(content)
