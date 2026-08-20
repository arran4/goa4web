//go:build !sqlite && !sqlite3

package db

import (
	"testing"
)

func TestNewForDriverStub(t *testing.T) {
	querierMySQL := NewForDriver(nil, "mysql")
	if _, ok := querierMySQL.(*Queries); !ok {
		t.Errorf("expected *Queries for mysql driver, got %T", querierMySQL)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic when requesting sqlite driver without sqlite build tags, but did not panic")
		}
	}()

	_ = NewForDriver(nil, "sqlite3")
}
