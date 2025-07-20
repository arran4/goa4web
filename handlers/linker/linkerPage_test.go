package linker

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
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
	mock.MatchExpectationsInOrder(false)
	defer sqldb.Close()

	queries := db.New(sqldb)

	mock.ExpectExec("INSERT INTO linker").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{"idlinker", "language_idlanguage", "users_idusers", "linkercategory_idlinkerCategory", "forumthread_idforumthread", "title", "url", "description", "listed", "username", "title_2"}).
		AddRow(1, 1, 1, 1, 0, "Foo", "http://foo", "Bar", time.Now(), "u", "c")
	mock.ExpectQuery("WITH RECURSIVE role_ids").WithArgs(int32(0), int32(1), sql.NullInt32{}).WillReturnRows(rows)

	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("foo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT IGNORE INTO linker_search").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("bar").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT IGNORE INTO linker_search").WithArgs(int32(1), int32(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE linker SET last_index").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	bus := eventbus.NewBus()
	eventbus.DefaultBus = bus
	defer func() { eventbus.DefaultBus = eventbus.NewBus() }()

	ctx, cancel := context.WithCancel(context.Background())
	go searchworker.Worker(ctx, bus, queries)

	evt := &eventbus.Event{Data: map[string]any{}}
	cd := &common.CoreData{}
	cd.SetEvent(evt)

	req := httptest.NewRequest("POST", "/admin/queue?qid=1", nil)
	ctxreq := context.WithValue(req.Context(), consts.KeyQueries, queries)
	ctxreq = context.WithValue(ctxreq, consts.KeyCoreData, cd)
	req = req.WithContext(ctxreq)
	rr := httptest.NewRecorder()
	ApproveTask.Action(rr, req)

	if err := eventbus.DefaultBus.Publish(*evt); err != nil {
		t.Fatalf("publish: %v", err)
	}

	bus.Shutdown(context.Background())
	time.Sleep(20 * time.Millisecond)
	cancel()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
