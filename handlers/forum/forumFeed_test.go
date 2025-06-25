package forum

import (
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"
)

func TestForumTopicFeed(t *testing.T) {
	rows := []*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{
		{
			Idforumthread:     1,
			Firstposttext:     sql.NullString{String: "hello world", Valid: true},
			Firstpostusername: sql.NullString{String: "bob", Valid: true},
			Firstpostwritten:  sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		},
	}
	r := httptest.NewRequest("GET", "http://example.com/forum/topic/1.rss", nil)
	feed := TopicFeed(r, "Test", 1, rows)
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
