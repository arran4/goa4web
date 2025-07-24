package linker

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/workers/searchworker"
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
	origCfg := config.AppRuntimeConfig
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })

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
	mock.ExpectQuery("SELECT l.idlinker").WithArgs(int32(0), int32(1), sqlmock.AnyArg()).WillReturnRows(rows)

	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("foo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO linker_search").WithArgs(int32(1), int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("bar").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT INTO linker_search").WithArgs(int32(1), int32(2), int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE linker SET last_index").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	bus := eventbus.NewBus()

	ctx, cancel := context.WithCancel(context.Background())
	go searchworker.Worker(ctx, bus, queries)

	req := httptest.NewRequest("POST", "/admin/queue?qid=1", nil)
	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(req.Context(), queries, common.WithConfig(config.AppRuntimeConfig))
	cd.SetEvent(evt)
	cd.SetEventTask(ApproveTask)
	ctxreq := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctxreq)
	rr := httptest.NewRecorder()
	ApproveTask.Action(rr, req)

	if err := bus.Publish(*evt); err != nil {
		t.Fatalf("publish: %v", err)
	}

	bus.Shutdown(context.Background())
	cancel()
	// Wait for the worker goroutine to exit before verifying expectations.
	time.Sleep(500 * time.Millisecond)
	err = nil
	for i := 0; i < 500; i++ {
		err = mock.ExpectationsWereMet()
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
