package externallink

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestReloadExternalLinkTask(t *testing.T) {
	key := "testkey"

	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Action with URL", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			qs.CreateExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}
			qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
				if arg.ID != 123 {
					t.Errorf("Expected ID 123, got %d", arg.ID)
				}
				if arg.CardTitle.String != "Test Title" {
					t.Errorf("Expected title 'Test Title', got '%s'", arg.CardTitle.String)
				}
				return nil
			}
			qs.UpdateExternalLinkImageCacheFn = func(ctx context.Context, arg db.UpdateExternalLinkImageCacheParams) error {
				return nil
			}
			qs.EnsureExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}

			link := "https://example.com/some/link"
			sig := sign.Sign("link:"+link, key)

			client := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test Title"/></head><body></body></html>`)),
					Header:     make(http.Header),
				}
			})

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key), common.WithHTTPClient(client))
			cd.SetEvent(&eventbus.TaskEvent{})

			u := "/?u=" + url.QueryEscape(link) + "&sig=" + sig
			req := httptest.NewRequest(http.MethodPost, u, nil)
			_ = req.ParseForm()

			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)

			if _, ok := res.(handlers.RedirectHandler); !ok {
				t.Fatalf("Expected handlers.RedirectHandler, got %T", res)
			}

			val := res.(handlers.RedirectHandler)
			if string(val) != u+"&msg=Reloading+Open+Graph+data+in+the+background..." {
				t.Errorf("Expected '%s', got '%s'", u+"&msg=Reloading+Open+Graph+data+in+the+background...", string(val))
			}

			// Verify task was set
			evtTask := cd.Event().Task
			if evtTask == nil {
				t.Errorf("Expected Task to be set in cd.Event()")
			}
		})

		t.Run("Action with ID", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			link := "https://example.com/some/link"
			qs.GetExternalLinkByIDFn = func(ctx context.Context, id int32) (*db.ExternalLink, error) {
				return &db.ExternalLink{ID: 123, Url: link}, nil
			}
			qs.CreateExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				// Should find existing or just update
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}
			qs.EnsureExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}
			qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
				if arg.ID != 123 {
					t.Errorf("Expected ID 123, got %d", arg.ID)
				}
				return nil
			}

			idStr := "123"
			sig := sign.Sign("link:"+idStr, key)

			client := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test Title"/></head><body></body></html>`)),
					Header:     make(http.Header),
				}
			})

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key), common.WithHTTPClient(client))
			cd.SetEvent(&eventbus.TaskEvent{})

			u := "/?id=" + idStr + "&sig=" + sig
			req := httptest.NewRequest(http.MethodPost, u, nil)
			_ = req.ParseForm()

			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)

			if _, ok := res.(handlers.RedirectHandler); !ok {
				t.Fatalf("Expected handlers.RedirectHandler, got %T", res)
			}

			// Verify task was set
			evtTask := cd.Event().Task
			if evtTask == nil {
				t.Errorf("Expected Task to be set in cd.Event()")
			}
		})
	})

	t.Run("Unhappy Path", func(t *testing.T) {
		t.Run("Invalid Signature", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key))

			link := "https://example.com"
			u := "/?u=" + url.QueryEscape(link) + "&sig=invalid"
			req := httptest.NewRequest(http.MethodPost, u, nil)
			_ = req.ParseForm()
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)
			if _, ok := res.(error); !ok {
				t.Errorf("Expected error, got %T", res)
			}
		})

		t.Run("Missing LinkSignKey", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			cd := common.NewCoreData(context.Background(), qs, nil) // No key

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)
			if err, ok := res.(error); !ok || err.Error() != "invalid link: Forbidden" {
				t.Errorf("Expected 'invalid link: Forbidden', got %v", res)
			}
		})
	})
}

func TestPrefetchExternalLinkTask(t *testing.T) {
	t.Run("Happy Path - Action with valid URL", func(t *testing.T) {
		qs := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(context.Background(), qs, nil)
		cd.SetEvent(&eventbus.TaskEvent{})

		req := httptest.NewRequest(http.MethodPost, "/prefetch?url="+url.QueryEscape("https://example.com/page?utm_source=x&id=1"), nil)
		_ = req.ParseForm()
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		rec := httptest.NewRecorder()

		res := prefetchExternalLinkTask.Action(rec, req)
		val, ok := res.(handlers.TextByteWriter)
		if !ok || string(val) != "ok" {
			t.Fatalf("expected TextByteWriter('ok'), got %v", res)
		}

		evtTask := cd.Event().Task
		prefetchTask, ok := evtTask.(PrefetchExternalLinkTask)
		if !ok {
			t.Fatalf("expected PrefetchExternalLinkTask, got %T", evtTask)
		}
		if prefetchTask.URL != "https://example.com/page?id=1" {
			t.Errorf("expected canonical URL 'https://example.com/page?id=1', got '%s'", prefetchTask.URL)
		}
	})

	t.Run("Unhappy Path - Missing URL does not enqueue background work", func(t *testing.T) {
		qs := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(context.Background(), qs, nil)
		cd.SetEvent(&eventbus.TaskEvent{})

		req := httptest.NewRequest(http.MethodPost, "/prefetch", nil)
		_ = req.ParseForm()
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		rec := httptest.NewRecorder()

		res := prefetchExternalLinkTask.Action(rec, req)
		val, ok := res.(handlers.TextByteWriter)
		if !ok || string(val) != "missing url" {
			t.Fatalf("expected TextByteWriter('missing url'), got %v", res)
		}

		evtTask := cd.Event().Task
		if _, ok := evtTask.(tasks.BackgroundTasker); ok {
			t.Errorf("expected missing URL to NOT attach a BackgroundTasker, got %T", evtTask)
		}
	})
}

func TestExternalLinkBackgroundTask_ExecutesAndPreservesHTTPClient(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	var metadataUpdated bool
	qs.EnsureExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
		return db.FakeSQLResult{LastInsertIDValue: 456}, nil
	}
	qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
		if arg.ID != 456 {
			t.Errorf("expected ID 456, got %d", arg.ID)
		}
		if arg.CardTitle.String != "OG Custom Title" {
			t.Errorf("expected title 'OG Custom Title', got '%s'", arg.CardTitle.String)
		}
		metadataUpdated = true
		return nil
	}

	var clientCalled bool
	client := NewTestClient(func(req *http.Request) *http.Response {
		clientCalled = true
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="OG Custom Title"/><meta property="og:description" content="OG Custom Desc"/></head><body></body></html>`)),
			Header:     make(http.Header),
		}
	})

	cd := common.NewCoreData(context.Background(), qs, nil, common.WithHTTPClient(client))
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	task := ReloadExternalLinkTask{
		TaskString: "admin:externallink:reload",
		URL:        "https://example.com/custom/article",
	}

	followUp, err := task.BackgroundTask(ctx, qs)
	if err != nil {
		t.Fatalf("BackgroundTask failed: %v", err)
	}
	if followUp != nil {
		t.Errorf("expected nil follow-up task, got %v", followUp)
	}
	if !clientCalled {
		t.Errorf("expected custom HTTP client to be called")
	}
	if !metadataUpdated {
		t.Errorf("expected UpdateExternalLinkMetadata to be called")
	}
}

func TestExternalLinkBackgroundTask_EmptyURLNoop(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	reloadTask := ReloadExternalLinkTask{URL: ""}
	if res, err := reloadTask.BackgroundTask(context.Background(), qs); err != nil || res != nil {
		t.Errorf("expected (nil, nil) for empty reload task, got (%v, %v)", res, err)
	}

	prefetchTask := PrefetchExternalLinkTask{URL: ""}
	if res, err := prefetchTask.BackgroundTask(context.Background(), qs); err != nil || res != nil {
		t.Errorf("expected (nil, nil) for empty prefetch task, got (%v, %v)", res, err)
	}
}
