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
	tests := []struct {
		name   string
		setup  func(sqlmock.Sqlmock)
		allow  bool
		cdInit func(db.Querier) *common.CoreData
	}{
		{
			name: "no grants",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)
			},
			allow: false,
			cdInit: func(q db.Querier) *common.CoreData {
				return common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			},
		},
		{
			name: "global grant",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
			},
			allow: true,
			cdInit: func(q db.Querier) *common.CoreData {
				return common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			},
		},
		{
			name: "section grant",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
			},
			allow: true,
			cdInit: func(q db.Querier) *common.CoreData {
				return common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer conn.Close()

			queries := db.New(conn)
			cd := tt.cdInit(queries)
			tt.setup(mock)

			if common.CanSearch(cd, "news") != tt.allow {
				t.Fatalf("expected allow=%v", tt.allow)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}
