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
	"github.com/arran4/goa4web/handlers/handlertest"
)

func TestNewsSearchFiltersUnauthorized(t *testing.T) {
	cd, mock, cleanup := handlertest.NewCoreData(t, context.Background())
	defer cleanup()

	firstRows := sqlmock.NewRows([]string{"site_news_id"}).AddRow(1).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT cs.site_news_id")).
		WithArgs(int32(1), sql.NullString{String: "foo", Valid: true}, int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(firstRows)

	newsRows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id",
		"language_id", "users_idusers", "news", "occurred", "timezone",
		"Comments",
	}).AddRow("bob", 1, 1, 0, 1, 1, "text", time.Unix(0, 0), time.Local.String(), 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.username AS writerName")).
		WithArgs(int32(1), int32(1), int32(2), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(newsRows)

	form := url.Values{"searchwords": {"foo"}}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	req = req.WithContext(ctx)
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
