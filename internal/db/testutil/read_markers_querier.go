package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// ReadMarkersQuerier implements read marker db.Querier methods for tests.
type ReadMarkersQuerier struct {
	*BaseQuerier
	Marker   *db.GetContentReadMarkerRow
	Upserted []db.UpsertContentReadMarkerParams
}

// NewReadMarkersQuerier returns a read marker querier stub.
func NewReadMarkersQuerier(t testing.TB) *ReadMarkersQuerier {
	t.Helper()
	return &ReadMarkersQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *ReadMarkersQuerier) UpsertContentReadMarker(ctx context.Context, arg db.UpsertContentReadMarkerParams) error {
	q.Upserted = append(q.Upserted, arg)
	return nil
}

func (q *ReadMarkersQuerier) GetContentReadMarker(ctx context.Context, arg db.GetContentReadMarkerParams) (*db.GetContentReadMarkerRow, error) {
	if q.Marker == nil {
		return nil, sql.ErrNoRows
	}
	return q.Marker, nil
}
