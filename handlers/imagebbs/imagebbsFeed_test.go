package imagebbs

import (
	"database/sql"
	"github.com/arran4/goa4web/internal/db"
	images "github.com/arran4/goa4web/internal/images"
	"net/http/httptest"
	"testing"
	"time"
)

func TestImagebbsFeed(t *testing.T) {
	rows := []*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow{
		{
			Idimagepost:   1,
			ForumthreadID: 2,
			Description:   sql.NullString{String: "hello", Valid: true},
			Username:      sql.NullString{String: "bob", Valid: true},
			Posted:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		},
	}
	r := httptest.NewRequest("GET", "http://example.com/imagebbs/board/1.rss", nil)
	feed := imagebbsFeed(r, "Test", 1, rows, images.NewSigner("k"))
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
