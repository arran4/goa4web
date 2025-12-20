package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// ThreadReplyQuerier implements thread reply checks for tests.
type ThreadReplyQuerier struct {
	*BaseQuerier
	Thread *db.Forumthread
	Err    error
}

// NewThreadReplyQuerier returns a thread reply querier stub.
func NewThreadReplyQuerier(t testing.TB) *ThreadReplyQuerier {
	t.Helper()
	return &ThreadReplyQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *ThreadReplyQuerier) GetThreadBySectionThreadIDForReplier(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
	if q.Err != nil {
		return nil, q.Err
	}
	return q.Thread, nil
}
