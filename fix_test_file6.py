import sys

with open("handlers/forum/forum_post_append_test.go", "r") as f:
    content = f.read()

bad = """	if m.GetCommentByIdFn != nil {
        return m.GetCommentByIdFn(ctx, id)
    }
    return m.QuerierStub.GetCommentById(ctx, id)"""

good = """	if m.GetCommentByIdFn != nil {
        return m.GetCommentByIdFn(ctx, id)
    }
    return &db.Comment{Idcomments: id}, nil"""

content = content.replace(bad, good)

with open("handlers/forum/forum_post_append_test.go", "w") as f:
    f.write(content)
