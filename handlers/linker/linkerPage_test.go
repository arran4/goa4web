package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func TestLinkerFeed(t *testing.T) {
	rows := []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow{
		{
			Idlinker:       1,
			Title:          sql.NullString{String: "Example", Valid: true},
			Url:            sql.NullString{String: "http://example.com", Valid: true},
			Description:    sql.NullString{String: "desc", Valid: true},
			Listed:         sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Posterusername: sql.NullString{String: "bob", Valid: true},
		},
	}
	r := httptest.NewRequest("GET", "http://example.com/linker/rss", nil)
	feed := linkerFeed(r, rows)
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item got %d", len(feed.Items))
	}
	if feed.Items[0].Link.Href != "http://example.com" {
		t.Errorf("unexpected link %s", feed.Items[0].Link.Href)
	}
	if feed.Items[0].Title != "Example" {
		t.Errorf("unexpected title %s", feed.Items[0].Title)
	}
}
func TestLinkerApproveAddsToSearch(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	queries := db.New(sqldb)

	// Approve item from queue id 1
	mock.ExpectExec("INSERT INTO linker").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{"idlinker", "language_idlanguage", "users_idusers", "linkercategory_idlinkerCategory", "forumthread_idforumthread", "title", "url", "description", "listed", "username", "title_2"}).
		AddRow(1, 1, 1, 1, 0, "Foo", "http://foo", "Bar", time.Now(), "u", "c")
	mock.ExpectQuery("SELECT l.idlinker").WithArgs(int32(1)).WillReturnRows(rows)

	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("foo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT IGNORE INTO linker_search").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("bar").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT IGNORE INTO linker_search").WithArgs(int32(1), int32(2)).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/admin/queue?qid=1", nil)
	ctx := context.WithValue(req.Context(), handlers.KeyQueries, queries)
	ctx = context.WithValue(ctx, handlers.KeyCoreData, &corecommon.CoreData{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	AdminQueueApproveActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
