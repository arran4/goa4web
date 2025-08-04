package search

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCanSearch(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, common.WithConfig(config.NewRuntimeConfig()))

	// No grants
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)
	if common.CanSearch(cd, "news") {
		t.Fatalf("expected false")
	}

	// Global grant only
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	if !common.CanSearch(cd, "news") {
		t.Fatalf("expected true with global grant")
	}

	// Grant present for section
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	if !common.CanSearch(cd, "news") {
		t.Fatalf("expected true with section grant")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
