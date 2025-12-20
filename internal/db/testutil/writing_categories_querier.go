package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// WritingCategoriesQuerier implements writing category list queries for tests.
type WritingCategoriesQuerier struct {
	*BaseQuerier
	Categories []*db.WritingCategory
}

// NewWritingCategoriesQuerier returns a writing categories querier stub.
func NewWritingCategoriesQuerier(t testing.TB) *WritingCategoriesQuerier {
	t.Helper()
	return &WritingCategoriesQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *WritingCategoriesQuerier) ListWritingCategoriesForLister(ctx context.Context, arg db.ListWritingCategoriesForListerParams) ([]*db.WritingCategory, error) {
	return q.Categories, nil
}
