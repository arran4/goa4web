package goa4web

import (
	"context"
	"encoding/xml"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBlogsRssPageWritesRSS(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := New(db)

	mock.ExpectQuery(regexp.QuoteMeta(getUserByUsername)).
		WithArgs("bob").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "passwd", "username"}).
			AddRow(1, "e", "p", "bob"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(15), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"idblogs", "forumthread_idforumthread", "users_idusers", "language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)"}).
			AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), "bob", 0))

	req := httptest.NewRequest("GET", "http://example.com/blogs/rss?rss=bob", nil)
	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	blogsRssPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/rss+xml" {
		t.Errorf("Content-Type=%q", ct)
	}

	var v struct{ XMLName xml.Name }
	if err := xml.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("xml parse: %v", err)
	}
	if v.XMLName.Local != "rss" {
		t.Errorf("expected root rss got %s", v.XMLName.Local)
	}
}
