package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// BlogQuerier implements blog list queries for tests.
type BlogQuerier struct {
	*BaseQuerier
	Bloggers            []*db.ListBloggersForListerRow
	Writers             []*db.ListWritersForListerRow
	BlogEntries         []*db.ListBlogEntriesForListerRow
	BlogEntriesByAuthor []*db.ListBlogEntriesByAuthorForListerRow
}

// NewBlogQuerier returns a blog querier stub.
func NewBlogQuerier(t testing.TB) *BlogQuerier {
	t.Helper()
	return &BlogQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *BlogQuerier) ListBloggersForLister(ctx context.Context, arg db.ListBloggersForListerParams) ([]*db.ListBloggersForListerRow, error) {
	return q.Bloggers, nil
}

func (q *BlogQuerier) ListWritersForLister(ctx context.Context, arg db.ListWritersForListerParams) ([]*db.ListWritersForListerRow, error) {
	return q.Writers, nil
}

func (q *BlogQuerier) ListBlogEntriesForLister(ctx context.Context, arg db.ListBlogEntriesForListerParams) ([]*db.ListBlogEntriesForListerRow, error) {
	return q.BlogEntries, nil
}

func (q *BlogQuerier) ListBlogEntriesByAuthorForLister(ctx context.Context, arg db.ListBlogEntriesByAuthorForListerParams) ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
	return q.BlogEntriesByAuthor, nil
}
