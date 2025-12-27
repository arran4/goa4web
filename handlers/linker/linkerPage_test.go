package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/workers/searchworker"
)

func linkerItemForSearch() *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow {
	return &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:          1,
		LanguageID:  sql.NullInt32{Int32: 1, Valid: true},
		AuthorID:    1,
		CategoryID:  sql.NullInt32{Int32: 1, Valid: true},
		ThreadID:    0,
		Title:       sql.NullString{String: "Foo", Valid: true},
		Url:         sql.NullString{String: "http://foo", Valid: true},
		Description: sql.NullString{String: "Bar", Valid: true},
		Listed:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Username:    sql.NullString{String: "u", Valid: true},
		Title_2:     sql.NullString{String: "c", Valid: true},
	}
}

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
	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: linkerItemForSearch(),
		AdminInsertQueuedLinkFromQueueReturn:                                    1,
	}

	bus := eventbus.NewBus()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go searchworker.Worker(ctx, bus, queries)
	time.Sleep(10 * time.Millisecond)

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

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(queries.SystemSetLinkerLastIndexCalls) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := bus.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if len(queries.AdminInsertQueuedLinkFromQueueCalls) != 1 || queries.AdminInsertQueuedLinkFromQueueCalls[0] != 1 {
		t.Fatalf("unexpected queued link insert calls: %+v", queries.AdminInsertQueuedLinkFromQueueCalls)
	}
	if len(queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls) != 1 {
		t.Fatalf("expected linker fetch, got %d", len(queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls))
	}
	if len(queries.SystemCreateSearchWordCalls) != 2 {
		t.Fatalf("expected two search words, got %d", len(queries.SystemCreateSearchWordCalls))
	}
	if len(queries.SystemAddToLinkerSearchCalls) != 2 {
		t.Fatalf("expected two linker search inserts, got %d", len(queries.SystemAddToLinkerSearchCalls))
	}
	if len(queries.SystemSetLinkerLastIndexCalls) != 1 || queries.SystemSetLinkerLastIndexCalls[0] != 1 {
		t.Fatalf("unexpected last index calls: %+v", queries.SystemSetLinkerLastIndexCalls)
	}
}
