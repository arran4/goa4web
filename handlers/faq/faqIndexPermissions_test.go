package faq

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCustomFAQIndexAsk(t *testing.T) {
	req := httptest.NewRequest("GET", "/faq", nil)

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "faq", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomFAQIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Ask") {
		t.Errorf("expected ask item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomFAQIndexAskDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/faq", nil)

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "faq", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	CustomFAQIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "Ask") {
		t.Errorf("unexpected ask item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
