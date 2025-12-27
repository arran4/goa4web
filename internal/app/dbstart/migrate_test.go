package dbstart

import (
	"context"
	"database/sql"
	"testing"

	"testing/fstest"
)

func TestApply(t *testing.T) {
	mfs := fstest.MapFS{
		"0002.mysql.sql": {Data: []byte("CREATE TABLE t (id int);")},
	}

	fq := &fakeQuerier{scanErr: sql.ErrNoRows}

	if err := applyWithQuerier(context.Background(), fq, mfs, false, "mysql"); err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(fq.execLog) != 4 {
		t.Fatalf("expected 4 exec calls, got %d", len(fq.execLog))
	}
	if got := fq.execLog[0]; got.query != "CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)" {
		t.Fatalf("unexpected first exec %q", got.query)
	}
	if got := fq.execLog[len(fq.execLog)-1]; got.query != "UPDATE schema_version SET version = ?" || got.args[0] != 2 {
		t.Fatalf("unexpected final exec %+v", got)
	}
}
