package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"sort"
	"sync"
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
	cd := &common.CoreData{ImageSignKey: "test-key",
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
	linkerID := int64(7)
	queries := newLinkerIndexRecorder(linkerID, &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:          int32(linkerID),
		Title:       sql.NullString{String: "Foo", Valid: true},
		Description: sql.NullString{String: "Bar baz", Valid: true},
		Username:    sql.NullString{String: "alice", Valid: true},
	})

	bus := eventbus.NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go searchworker.Worker(ctx, bus, queries)

	req := httptest.NewRequest("POST", "/admin/queue?qid=3", nil)
	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	cd.SetEvent(evt)
	cd.SetEventTask(AdminApproveTask)
	ctxreq := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctxreq)
	rr := httptest.NewRecorder()
	if err := AdminApproveTask.Action(rr, req); err != nil {
		t.Fatalf("action: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if data, ok := evt.Data[searchworker.EventKey].(searchworker.IndexEventData); !ok {
		t.Fatalf("expected index data in event")
	} else {
		if data.Type != searchworker.TypeLinker {
			t.Fatalf("index type = %q, want %q", data.Type, searchworker.TypeLinker)
		}
		if data.ID != int32(linkerID) {
			t.Fatalf("index id = %d, want %d", data.ID, linkerID)
		}
		if data.Text != "Foo Bar baz" {
			t.Fatalf("index text = %q", data.Text)
		}
	}

	if err := bus.Publish(*evt); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case <-queries.indexed:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for index updates")
	}

	cancel()
	_ = bus.Shutdown(context.Background())

	if got := queries.queuedIDs(); len(got) != 1 || got[0] != 3 {
		t.Fatalf("queued IDs = %v", got)
	}

	if got := queries.createdWords(); !slicesEqual(got, []string{"bar", "baz", "foo"}) {
		t.Fatalf("created words = %v", got)
	}

	calls := queries.searchCalls()
	if len(calls) != 3 {
		t.Fatalf("search calls = %v", calls)
	}
	for word, count := range map[string]int32{"foo": 1, "bar": 1, "baz": 1} {
		if params, ok := calls[word]; !ok {
			t.Fatalf("missing search call for %s", word)
		} else if params.LinkerID != int32(linkerID) || params.WordCount != count {
			t.Fatalf("search params for %s = %+v", word, params)
		}
	}

	if got := queries.lastIndexIDs(); len(got) != 1 || got[0] != int32(linkerID) {
		t.Fatalf("last index updates = %v", got)
	}
}

type linkerIndexRecorder struct {
	db.Querier
	mu             sync.Mutex
	queued         []int32
	linkID         int64
	link           *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
	words          []string
	wordIDs        map[string]int64
	wordNames      map[int64]string
	searchCallsLog []db.SystemAddToLinkerSearchParams
	lastIndex      []int32
	indexed        chan struct{}
}

func newLinkerIndexRecorder(linkID int64, link *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow) *linkerIndexRecorder {
	return &linkerIndexRecorder{
		linkID:    linkID,
		link:      link,
		wordIDs:   map[string]int64{},
		wordNames: map[int64]string{},
		indexed:   make(chan struct{}, 1),
	}
}

func (l *linkerIndexRecorder) AdminInsertQueuedLinkFromQueue(_ context.Context, id int32) (int64, error) {
	l.mu.Lock()
	l.queued = append(l.queued, id)
	l.mu.Unlock()
	return l.linkID, nil
}

func (l *linkerIndexRecorder) GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(context.Context, db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	return l.link, nil
}

func (l *linkerIndexRecorder) SystemCreateSearchWord(_ context.Context, word string) (int64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if id, ok := l.wordIDs[word]; ok {
		return id, nil
	}
	id := int64(len(l.wordIDs) + 1)
	l.wordIDs[word] = id
	l.wordNames[id] = word
	l.words = append(l.words, word)
	return id, nil
}

func (l *linkerIndexRecorder) SystemAddToLinkerSearch(_ context.Context, params db.SystemAddToLinkerSearchParams) error {
	l.mu.Lock()
	l.searchCallsLog = append(l.searchCallsLog, params)
	l.mu.Unlock()
	return nil
}

func (l *linkerIndexRecorder) SystemSetLinkerLastIndex(_ context.Context, linkerID int32) error {
	l.mu.Lock()
	l.lastIndex = append(l.lastIndex, linkerID)
	l.mu.Unlock()
	select {
	case l.indexed <- struct{}{}:
	default:
	}
	return nil
}

func (l *linkerIndexRecorder) queuedIDs() []int32 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]int32(nil), l.queued...)
}

func (l *linkerIndexRecorder) createdWords() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	words := append([]string(nil), l.words...)
	sort.Strings(words)
	return words
}

func (l *linkerIndexRecorder) searchCalls() map[string]db.SystemAddToLinkerSearchParams {
	l.mu.Lock()
	defer l.mu.Unlock()
	calls := map[string]db.SystemAddToLinkerSearchParams{}
	for _, params := range l.searchCallsLog {
		if word, ok := l.wordNames[int64(params.SearchwordlistIdsearchwordlist)]; ok {
			calls[word] = params
		}
	}
	return calls
}

func (l *linkerIndexRecorder) lastIndexIDs() []int32 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]int32(nil), l.lastIndex...)
}

func slicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
