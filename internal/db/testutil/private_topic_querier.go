package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// PrivateTopicQuerier implements private topic creation related queries for tests.
type PrivateTopicQuerier struct {
	*BaseQuerier
	Users          map[int32]*db.SystemGetUserByIDRow
	CreatedTopic   db.CreateForumTopicForPosterParams
	CreatedTopicID int64
	Grants         []db.SystemCreateGrantParams
}

// NewPrivateTopicQuerier returns a private topic querier stub.
func NewPrivateTopicQuerier(t testing.TB) *PrivateTopicQuerier {
	t.Helper()
	return &PrivateTopicQuerier{
		BaseQuerier: NewBaseQuerier(t),
		Users:       map[int32]*db.SystemGetUserByIDRow{},
	}
}

func (q *PrivateTopicQuerier) SystemGetUserByID(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
	if row, ok := q.Users[idusers]; ok {
		return row, nil
	}
	return nil, sql.ErrNoRows
}

func (q *PrivateTopicQuerier) CreateForumTopicForPoster(ctx context.Context, arg db.CreateForumTopicForPosterParams) (int64, error) {
	q.CreatedTopic = arg
	return q.CreatedTopicID, nil
}

func (q *PrivateTopicQuerier) SystemCreateGrant(ctx context.Context, arg db.SystemCreateGrantParams) (int64, error) {
	q.Grants = append(q.Grants, arg)
	return 1, nil
}
