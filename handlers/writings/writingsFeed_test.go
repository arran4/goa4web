package writings

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"regexp"
)

func TestWritingsFeed(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)

	req := httptest.NewRequest("GET", "http://example.com/writings/rss", nil)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSiteTitle("Site"))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mock.ExpectQuery(regexp.QuoteMeta("-- name: GetPublicWritings :many")).
		WillReturnRows(sqlmock.NewRows([]string{
			"idwriting", "users_idusers", "forumthread_id", "language_id",
			"writing_category_id", "title", "published", "timezone",
			"writing", "abstract", "private", "deleted_at", "last_index",
		}).
			AddRow(1, 1, 1, 1, 1, "Title", time.Now(), "UTC", "Content", "Abstract", false, nil, nil))

	mock.ExpectQuery(regexp.QuoteMeta("-- name: SystemCheckGrant :one")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	feed, err := feedGen(req, cd)
	if err != nil {
		t.Fatalf("feedGen: %v", err)
	}

	if feed.Title != "Site - Latest writings" {
		t.Errorf("feed title incorrect: %s", feed.Title)
	}
	if len(feed.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(feed.Items))
	}
}
