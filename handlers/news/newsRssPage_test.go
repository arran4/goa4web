package news

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewsRssPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	req := httptest.NewRequest("GET", "http://example.com/news/rss", nil)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSiteTitle("Site"))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	// Mock LatestNews query
	// GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending
	mock.ExpectQuery(regexp.QuoteMeta("-- name: GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many")).
		WillReturnRows(sqlmock.NewRows([]string{
			"writerName", "writerId", "idsiteNews", "forumthread_id", "language_id",
			"users_idusers", "news", "occurred", "timezone", "Comments",
		}).
			AddRow("Writer", 1, 1, 1, 1, 1, "News", time.Now(), "UTC", 0))

	// Mock HasGrant
	// Permissions() -> returns empty
	// HasAdminRole -> false
	// SystemCheckGrant
	mock.ExpectQuery(regexp.QuoteMeta("-- name: SystemCheckGrant :one")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	w := httptest.NewRecorder()
	NewsRssPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code: %d", w.Code)
	}

	var v struct {
		XMLName xml.Name
		Channel struct {
			Title string `xml:"title"`
		} `xml:"channel"`
	}
	if err := xml.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("xml parse: %v", err)
	}
	if v.Channel.Title != "Site - News feed" {
		t.Errorf("expected title 'Site - News feed' got %q", v.Channel.Title)
	}
}
