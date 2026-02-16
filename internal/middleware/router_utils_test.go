package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMiddlewareChain(t *testing.T) {
	// Helper to create middleware that records its execution
	createMW := func(name string, record *[]string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				*record = append(*record, name+"_start")
				next.ServeHTTP(w, r)
				*record = append(*record, name+"_end")
			})
		}
	}

	t.Run("Empty chain", func(t *testing.T) {
		handlerCalled := false
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		chain := NewMiddlewareChain()
		wrapped := chain.Wrap(finalHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if !handlerCalled {
			t.Error("Final handler should be called")
		}
	})

	t.Run("Single middleware", func(t *testing.T) {
		var record []string
		mw1 := createMW("mw1", &record)

		handlerCalled := false
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			record = append(record, "handler")
		})

		chain := NewMiddlewareChain(mw1)
		wrapped := chain.Wrap(finalHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if !handlerCalled {
			t.Error("Final handler should be called")
		}

		expected := []string{"mw1_start", "handler", "mw1_end"}
		if len(record) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, record)
		}
		for i := range expected {
			if i < len(record) && record[i] != expected[i] {
				t.Errorf("Expected %v, got %v", expected, record)
			}
		}
	})

	t.Run("Multiple middleware order", func(t *testing.T) {
		var record []string
		mw1 := createMW("mw1", &record)
		mw2 := createMW("mw2", &record)
		mw3 := createMW("mw3", &record)

		handlerCalled := false
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			record = append(record, "handler")
		})

		// Expected order: mw1 -> mw2 -> mw3 -> handler
		chain := NewMiddlewareChain(mw1, mw2, mw3)
		wrapped := chain.Wrap(finalHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if !handlerCalled {
			t.Error("Final handler should be called")
		}

		expected := []string{
			"mw1_start",
			"mw2_start",
			"mw3_start",
			"handler",
			"mw3_end",
			"mw2_end",
			"mw1_end",
		}

		if len(record) != len(expected) {
			t.Errorf("Length mismatch: Expected %d, got %d. Full expected: %v, got: %v", len(expected), len(record), expected, record)
		} else {
			for i := range expected {
				if record[i] != expected[i] {
					t.Errorf("At index %d: Expected %s, got %s. Full expected: %v, got: %v", i, expected[i], record[i], expected, record)
				}
			}
		}
	})

	t.Run("RouterWrapperFunc", func(t *testing.T) {
		called := false
		wrapper := RouterWrapperFunc(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				h.ServeHTTP(w, r)
			})
		})

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		wrapped := wrapper.Wrap(finalHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if !called {
			t.Error("Wrapper function should be called")
		}
	})
}
