package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserMonthlyUsageCounts_LinkerJoin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	q := New(db)

	// Tables in order of processing in UserMonthlyUsageCounts
	tables := []string{"blogs", "site_news", "comments", "imagepost", "linker", "writing"}

	for _, table := range tables {
		var expectedJoin string
		if table == "linker" {
			expectedJoin = "JOIN users u ON t.author_id = u.idusers"
		} else {
			expectedJoin = "JOIN users u ON t.users_idusers = u.idusers"
		}

		// Regex to match the query.
		regexStr := `SELECT .* FROM ` + table + ` t ` + regexp.QuoteMeta(expectedJoin) + ` WHERE .*`

		mock.ExpectQuery(regexStr).
			WithArgs(int32(2024)).
			WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "year", "month", "count"}))
	}

	_, err = q.UserMonthlyUsageCounts(context.Background(), 2024)
	if err != nil {
		t.Errorf("UserMonthlyUsageCounts failed: %v", err)
	}

	// Make sure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
