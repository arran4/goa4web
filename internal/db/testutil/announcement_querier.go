package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// AnnouncementQuerier implements announcement queries for tests.
type AnnouncementQuerier struct {
	*BaseQuerier
	Announcement *db.SiteAnnouncement
	Err          error
}

// NewAnnouncementQuerier returns an announcement querier stub.
func NewAnnouncementQuerier(t testing.TB) *AnnouncementQuerier {
	t.Helper()
	return &AnnouncementQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *AnnouncementQuerier) GetLatestAnnouncementByNewsID(ctx context.Context, id int32) (*db.SiteAnnouncement, error) {
	if q.Err != nil {
		return nil, q.Err
	}
	return q.Announcement, nil
}
