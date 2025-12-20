package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// WritingsQuerier implements writing list queries for tests.
type WritingsQuerier struct {
	*BaseQuerier
	PublicRows           []*db.ListPublicWritingsInCategoryForListerRow
	PublicRowsByCategory map[int32][]*db.ListPublicWritingsInCategoryForListerRow
	Writings             []*db.Writing
}

// NewWritingsQuerier returns a writings querier stub.
func NewWritingsQuerier(t testing.TB) *WritingsQuerier {
	t.Helper()
	return &WritingsQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *WritingsQuerier) ListPublicWritingsInCategoryForLister(ctx context.Context, arg db.ListPublicWritingsInCategoryForListerParams) ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
	if q.PublicRowsByCategory != nil {
		return q.PublicRowsByCategory[arg.WritingCategoryID], nil
	}
	return q.PublicRows, nil
}

func (q *WritingsQuerier) GetPublicWritings(ctx context.Context, arg db.GetPublicWritingsParams) ([]*db.Writing, error) {
	return q.Writings, nil
}
