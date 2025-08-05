package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
)

func TestQueries_InsertFAQQuestionForWriter(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(insertFAQQuestionForWriter)).
		WithArgs(sql.NullString{String: "q", Valid: true}, sql.NullString{String: "a", Valid: true}, int32(1), int32(2), int32(1), sql.NullInt32{Int32: 2, Valid: true}, int32(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := q.InsertFAQQuestionForWriter(context.Background(), InsertFAQQuestionForWriterParams{
		Question:   sql.NullString{String: "q", Valid: true},
		Answer:     sql.NullString{String: "a", Valid: true},
		CategoryID: 1,
		WriterID:   2,
		LanguageID: 1,
		GranteeID:  sql.NullInt32{Int32: 2, Valid: true},
	}); err != nil {
		t.Fatalf("InsertFAQQuestionForWriter: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestFAQQueriesAllowGlobalGrants(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"createFAQQuestionForWriter", createFAQQuestionForWriter},
		{"insertFAQQuestionForWriter", insertFAQQuestionForWriter},
		{"insertFAQRevisionForUser", insertFAQRevisionForUser},
	}

	itemSub := "(g.item = 'question' OR g.item IS NULL)"

	for _, c := range cases {
		if !strings.Contains(c.query, itemSub) {
			t.Errorf("%s missing global item check", c.name)
		}
	}
}
