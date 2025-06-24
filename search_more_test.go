package goa4web

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubDB struct {
	word string
	err  error
}

func (s *stubDB) ExecContext(ctx context.Context, q string, args ...interface{}) (sql.Result, error) {
	if s.err != nil {
		return nil, s.err
	}
	if len(args) > 0 {
		s.word = args[0].(string)
	}
	return stubResult{}, nil
}
func (s *stubDB) PrepareContext(context.Context, string) (*sql.Stmt, error) { return nil, nil }
func (s *stubDB) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (s *stubDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row { return &sql.Row{} }

type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 1, nil }
func (stubResult) RowsAffected() (int64, error) { return 1, nil }

func TestSearchWordIdsFromText(t *testing.T) {
	db := &stubDB{}
	q := New(db)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	ids, redirect := SearchWordIdsFromText(rr, req, "Hello world Hello", q)
	if redirect {
		t.Fatalf("unexpected redirect")
	}
	if len(ids) != 2 {
		t.Fatalf("ids=%v", ids)
	}
	if db.word != "hello" && db.word != "world" {
		t.Fatalf("word %s", db.word)
	}
}

func TestSearchWordIdsFromTextError(t *testing.T) {
	db := &stubDB{err: errors.New("bad")}
	q := New(db)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	ids, redirect := SearchWordIdsFromText(rr, req, "bad", q)
	if ids != nil {
		t.Fatal("expected nil ids")
	}
	if !redirect {
		t.Fatal("expected redirect")
	}
	if rr.Result().StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("status %d", rr.Result().StatusCode)
	}
}
