package news

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestNewsSearchFiltersUnauthorized(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	firstRows := sqlmock.NewRows([]string{"site_news_id"}).AddRow(1).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT cs.site_news_id")).
		WithArgs(int32(1), sql.NullString{String: "foo", Valid: true}, int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(firstRows)

	newsRows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id",
		"language_idlanguage", "users_idusers", "news", "occurred",
		"Comments",
	}).AddRow("bob", 1, 1, 0, 1, 1, "text", time.Unix(0, 0), 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.username AS writerName")).
		WithArgs(int32(1), int32(1), int32(2), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(newsRows)

	form := url.Values{"searchwords": {"foo"}}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	req = req.WithContext(ctx)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	news, emptyWords, noResults, err := cd.SearchNews(req, 1)
	if err != nil {
		t.Fatalf("NewsSearch: %v", err)
	}
	if emptyWords || noResults {
		t.Fatalf("unexpected flags")
	}
	if len(news) != 1 {
		t.Fatalf("expected 1 result got %d", len(news))
	}
	if news[0].Idsitenews != 1 {
		t.Errorf("unexpected id %d", news[0].Idsitenews)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
