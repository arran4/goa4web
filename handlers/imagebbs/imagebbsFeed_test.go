package imagebbs

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
)

func TestImagebbsFeed(t *testing.T) {
	rows := []*db.ListImagePostsByBoardForListerRow{
		{
			Idimagepost:   1,
			ForumthreadID: 2,
			Description:   sql.NullString{String: "hello", Valid: true},
			Username:      sql.NullString{String: "bob", Valid: true},
			Posted:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		},
	}
	r := httptest.NewRequest("GET", "http://example.com/imagebbs/board/1.rss", nil)
	cd := &common.CoreData{ImageSignKey: "test-key",
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	feed := cd.ImageBBSFeed(r, "Test", 1, rows)
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item got %d", len(feed.Items))
	}
	if feed.Items[0].Link.Href != "/imagebbs/board/1/thread/2" {
		t.Errorf("unexpected link %s", feed.Items[0].Link.Href)
	}
	if feed.Items[0].Title != "hello" {
		t.Errorf("unexpected title %s", feed.Items[0].Title)
	}
}
