package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestMonthlyUsageCounts ensures that writing statistics are included in the
// monthly usage aggregation.
func TestMonthlyUsageCounts(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := New(sqlDB)

	startYear := int32(2024)
	expect := func(table, column string) {
		q := "SELECT YEAR(" + column + "), MONTH(" + column + "), COUNT(*) FROM " + table + " WHERE YEAR(" + column + ") >= ? GROUP BY YEAR(" + column + "), MONTH(" + column + ")"
		rows := sqlmock.NewRows([]string{"year", "month", "count"}).AddRow(int32(2024), int32(1), int64(1))
		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WithArgs(startYear).
			WillReturnRows(rows)
	}

	expect("blogs", "written")
	expect("site_news", "occurred")
	expect("comments", "written")
	expect("imagepost", "posted")
	expect("linker", "listed")
	expect("writing", "published")

	rows, err := queries.MonthlyUsageCounts(context.Background(), startYear)
	if err != nil {
		t.Fatalf("MonthlyUsageCounts: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Writings != 1 {
		t.Fatalf("expected writings=1, got %d", rows[0].Writings)
	}
}
