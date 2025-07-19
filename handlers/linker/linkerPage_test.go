package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	searchworker "github.com/arran4/goa4web/workers/searchworker"
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

	req := httptest.NewRequest("POST", "/admin/queue?qid=1", nil)
	evt := &eventbus.Event{}
	cd := &common.CoreData{}
	cd.SetEvent(evt)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	ApproveTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	data, ok := evt.Data[searchworker.EventKey].(searchworker.IndexEventData)
	if !ok {
		t.Fatalf("missing search event: %+v", evt.Data)
	}
	if data.ID != 1 {
		t.Errorf("id=%d", data.ID)
	}
	if data.Text != "Foo Bar" {
		t.Errorf("text=%q", data.Text)
	}
}
