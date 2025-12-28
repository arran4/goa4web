package db

import (
	"context"
	"database/sql"
	"testing"
)

func TestQuerierStub_InsertFAQQuestionForWriter(t *testing.T) {
	stub := &QuerierStub{
		InsertFAQQuestionForWriterResult: FakeSQLResult{
			LastInsertIDValue: 42,
			RowsAffectedValue: 1,
		},
	}

	params := InsertFAQQuestionForWriterParams{
		Question: sql.NullString{String: "Why?", Valid: true},
	}

	res, err := stub.InsertFAQQuestionForWriter(context.Background(), params)
	if err != nil {
		t.Fatalf("InsertFAQQuestionForWriter returned error: %v", err)
	}

	if len(stub.InsertFAQQuestionForWriterCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(stub.InsertFAQQuestionForWriterCalls))
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId returned error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected last insert ID 42, got %d", id)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected returned error: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected rows affected 1, got %d", affected)
	}
}
