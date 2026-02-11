package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestTaskEventMiddleware(t *testing.T) {
	bus := eventbus.NewBus()
	mw := NewTaskEventMiddleware(bus)
	successHandler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/admin/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ch := bus.Subscribe(eventbus.TaskMessageType)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	successHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		evt, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong type %T", msg)
		}
		named, ok := evt.Task.(tasks.Name)
		if !ok || named.Name() != "MISSING" || evt.Path != "/admin/p" {
			t.Fatalf("unexpected event %+v", evt)
		}
	default:
		t.Fatalf("expected event on success")
	}

	// non-admin path should not include /admin prefix
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe(eventbus.TaskMessageType)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	successHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		evt, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong type %T", msg)
		}
		if strings.Contains(evt.Path, "/admin") {
			t.Fatalf("unexpected admin path for %#v", evt)
		}
	default:
		t.Fatalf("expected event for non-admin path")
	}

	failureHandler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("fail"))
	}))
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe(eventbus.TaskMessageType)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	failureHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		t.Fatalf("did not expect event on failure, got %T", msg)
		t.Fatalf("did not expect event on failure")
	default:
	}

	// ensure handlers can attach event data
	itemHandler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["info"] = true
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe(eventbus.TaskMessageType)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	itemHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		evt, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong type %T", msg)
		}
		val, ok := evt.Data["info"].(bool)
		if evt.Data == nil || !ok || !val {
			t.Fatalf("missing data: %+v", evt)
		}
	default:
		t.Fatalf("expected event with data")
	}
}

type countingWriter struct {
	http.ResponseWriter
	headerCalls int
}

func (cw *countingWriter) WriteHeader(code int) {
	cw.headerCalls++
	cw.ResponseWriter.WriteHeader(code)
}

func TestStatusRecorderWriteHeaderOnce(t *testing.T) {
	cw := &countingWriter{ResponseWriter: httptest.NewRecorder()}
	sr := &statusRecorder{ResponseWriter: cw}
	sr.WriteHeader(http.StatusTeapot)
	sr.WriteHeader(http.StatusInternalServerError)
	if cw.headerCalls != 1 {
		t.Fatalf("expected 1 WriteHeader call, got %d", cw.headerCalls)
	}
	if sr.status != http.StatusTeapot {
		t.Fatalf("status=%d", sr.status)
	}
}

func TestTaskEventQueue(t *testing.T) {
	bus := eventbus.NewBus()
	mw := NewTaskEventMiddleware(bus)

	if err := bus.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown bus: %v", err)
	}

	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	handler.ServeHTTP(rec, req.WithContext(ctx))

	if len(mw.Events()) != 1 {
		t.Fatalf("expected queued event")
	}

	bus = eventbus.NewBus()
	mw.SetBus(bus)
	ch := bus.Subscribe(eventbus.TaskMessageType)
	mw.Flush(context.Background())

	select {
	case <-ch:
	default:
		t.Fatalf("expected flushed event")
	}
}

func TestTaskEventMiddleware_EventProvided(t *testing.T) {
	handler := NewTaskEventMiddleware(eventbus.NewBus()).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil || cd.Event() == nil {
			t.Fatalf("missing event")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestTaskEventMiddleware_NoCoreDataPanic(t *testing.T) {
	handler := NewTaskEventMiddleware(eventbus.NewBus()).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not run")
	}))

	req := httptest.NewRequest("GET", "/p", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	handler.ServeHTTP(rec, req)
}
