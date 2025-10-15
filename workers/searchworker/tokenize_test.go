package searchworker

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestIsAlphanumericOrPunctuation(t *testing.T) {
	cases := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'-', true},
		{'\'', true},
		{'!', false},
	}
	for _, c := range cases {
		if got := IsAlphanumericOrPunctuation(c.r); got != c.want {
			t.Errorf("%q got %v want %v", string(c.r), got, c.want)
		}
	}
}

func TestIsAlphanumericOrPunctuationExtra(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'ñ', true},
		{'?', false},
		{'-', true},
		{'é', true},
		{'.', false},
		{'世', true},
	}
	for _, tt := range tests {
		if got := IsAlphanumericOrPunctuation(tt.r); got != tt.want {
			t.Errorf("%q got %v want %v", string(tt.r), got, tt.want)
		}
	}
}

func TestBreakupTextToWords(t *testing.T) {
	in := "Hello, world! It's-me"
	words := BreakupTextToWords(in)
	want := []string{"Hello", "world", "It's-me"}
	if len(words) != len(want) {
		t.Fatalf("len=%d want %d", len(words), len(want))
	}
	for i := range words {
		if words[i] != want[i] {
			t.Errorf("index %d got %q want %q", i, words[i], want[i])
		}
	}
}

func TestBreakupTextToWordsEdge(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"..hi--there--foo", []string{"hi--there--foo"}},
		{"end.", []string{"end"}},
		{"it's foo", []string{"it's", "foo"}},
		{"---abc", []string{"---abc"}},
		{"Hello...world!!", []string{"Hello", "world"}},
		{"foo   bar", []string{"foo", "bar"}},
		{"こんにちは 世界", []string{"こんにちは", "世界"}},
	}
	for _, c := range cases {
		got := BreakupTextToWords(c.in)
		if len(got) != len(c.want) {
			t.Errorf("%q len=%d want %d", c.in, len(got), len(c.want))
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q index %d got %q want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

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
	sdb := &stubDB{}
	q := db.New(sdb)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	ids, redirect := SearchWordIdsFromText(rr, req, "Hello world Hello", q)
	if redirect {
		t.Fatalf("unexpected redirect")
	}
	if len(ids) != 2 {
		t.Fatalf("ids=%v", ids)
	}
	if sdb.word != "hello" && sdb.word != "world" {
		t.Fatalf("word %s", sdb.word)
	}
	if ids[0].Count == 0 || ids[1].Count == 0 {
		t.Fatalf("counts=%v", ids)
	}
}

func TestSearchWordIdsFromTextError(t *testing.T) {
	sdb := &stubDB{err: errors.New("bad")}
	q := db.New(sdb)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	ids, redirect := SearchWordIdsFromText(rr, req, "bad", q)
	if ids != nil {
		t.Fatal("expected nil ids")
	}
	if !redirect {
		t.Fatal("expected redirect")
	}
	if rr.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("status %d", rr.Result().StatusCode)
	}
}
