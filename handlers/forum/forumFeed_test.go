package forum

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

func TestForumTopicFeed(t *testing.T) {
	rows := []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{
		{
			Idforumthread:     1,
			Firstposttext:     sql.NullString{String: "hello world", Valid: true},
			Firstpostusername: sql.NullString{String: "bob", Valid: true},
			Firstpostuserid:   sql.NullInt32{Int32: 1, Valid: true},
			Firstpostwritten:  sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		},
	}
	r := httptest.NewRequest("GET", "http://example.com/forum/topic/1.rss", nil)
	cd := &common.CoreData{ImageSignKey: "test-key",
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	feed := TopicFeed(r, "Test", 1, rows, "/forum")
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item got %d", len(feed.Items))
	}
	if feed.Items[0].Link.Href != "/forum/topic/1/thread/1" {
		t.Errorf("unexpected link %s", feed.Items[0].Link.Href)
	}
	if feed.Items[0].Title != "hello world" {
		t.Errorf("unexpected title %s", feed.Items[0].Title)
	}
}
