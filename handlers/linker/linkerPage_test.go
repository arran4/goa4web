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
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/workers/searchworker"
)

func TestLinkerFeed(t *testing.T) {
	rows := []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow{
		{
			ID:             1,
			Title:          sql.NullString{String: "Example", Valid: true},
			Url:            sql.NullString{String: "http://example.com", Valid: true},
			Description:    sql.NullString{String: "desc", Valid: true},
			Listed:         sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Posterusername: sql.NullString{String: "bob", Valid: true},
		},
	}
	r := httptest.NewRequest("GET", "http://example.com/linker/rss", nil)
	cd := &common.CoreData{ImageSigner: imagesign.NewSigner(&config.RuntimeConfig{}, "k")}
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
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
	t.Skip("event bus worker requires long wait; skipping")

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	defer conn.Close()

	queries := db.New(conn)

	mock.ExpectExec("INSERT INTO linker").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{"id", "language_id", "author_id", "category_id", "thread_id", "title", "url", "description", "listed", "username", "title_2"}).
		AddRow(1, 1, 1, 1, 0, "Foo", "http://foo", "Bar", time.Now(), "u", "c")
	mock.ExpectQuery("SELECT l.id").WithArgs(int32(0), int32(1), sqlmock.AnyArg()).WillReturnRows(rows)

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
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	cd.SetEvent(evt)
	cd.SetEventTask(AdminApproveTask)
	ctxreq := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctxreq)
	rr := httptest.NewRecorder()
	AdminApproveTask.Action(rr, req)

	if err := bus.Publish(*evt); err != nil {
		t.Fatalf("publish: %v", err)
	}

	bus.Shutdown(context.Background())
	cancel()
	// Wait for the worker goroutine to exit before verifying expectations.
	time.Sleep(1 * time.Second)
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
