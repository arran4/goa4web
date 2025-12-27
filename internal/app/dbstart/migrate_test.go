package dbstart

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestApply(t *testing.T) {
	mfs := fstest.MapFS{
		"0002.mysql.sql": {Data: []byte("CREATE TABLE t (id int);")},
	}

	store := &fakeStore{version: 1, hasVersion: true}

	if err := apply(context.Background(), store, mfs, false, "mysql"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if store.version != 2 {
		t.Fatalf("expected schema version 2, got %d", store.version)
	}
	foundMigration := false
	for _, exec := range store.execs {
		if exec == "CREATE TABLE t (id int) []" {
			foundMigration = true
			break
		}
	}
	if !foundMigration {
		t.Fatalf("expected migration statement to be executed, got %v", store.execs)
	}
}
