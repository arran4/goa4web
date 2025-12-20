package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// NewsQuerier implements news listing queries for tests.
type NewsQuerier struct {
	*BaseQuerier
	Posts []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
}

// NewNewsQuerier returns a news querier stub.
func NewNewsQuerier(t testing.TB) *NewsQuerier {
	t.Helper()
	return &NewsQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *NewsQuerier) GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(ctx context.Context, arg db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams) ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	return q.Posts, nil
}
