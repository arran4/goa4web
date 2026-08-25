import sys

new_content = """func (m *mockQuerier) GetCommentByIdForUser(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
    if m.GetCommentByIdFn != nil {
        comment, err := m.GetCommentByIdFn(ctx, arg.ID)
        if err != nil || comment == nil {
            return nil, err
        }
        return &db.GetCommentByIdForUserRow{
            Idcomments: comment.Idcomments,
            ForumthreadID: comment.ForumthreadID,
            UsersIdusers: comment.UsersIdusers,
            Text: comment.Text,
        }, nil
    }
    return &db.GetCommentByIdForUserRow{Idcomments: arg.ID}, nil
}"""

with open("handlers/forum/forum_post_append_test.go", "r") as f:
    content = f.read()

content = content + "\n" + new_content

with open("handlers/forum/forum_post_append_test.go", "w") as f:
    f.write(content)
