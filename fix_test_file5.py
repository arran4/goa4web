import sys

with open("handlers/forum/forum_post_append_test.go", "r") as f:
    content = f.read()

content = content.replace("testhelpers.QuerierStub", "db.QuerierStub")

with open("handlers/forum/forum_post_append_test.go", "w") as f:
    f.write(content)
