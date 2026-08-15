package common

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestEnsurePrivateForumTopicSeeGrantCreatesMissingGrant(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cd := NewTestCoreData(t, db.New(conn))
	const userID int32 = 23

	mock.ExpectQuery("WITH role_ids AS \\(").
		WithArgs(
			userID,
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"see",
			sql.NullInt32{},
			sql.NullInt32{Int32: userID, Valid: true},
		).
		WillReturnRows(sqlmock.NewRows([]string{"1"}))

	mock.ExpectExec("INSERT INTO grants").
		WithArgs(
			sql.NullInt32{Int32: userID, Valid: true},
			sql.NullInt32{},
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"allow",
			sql.NullInt32{},
			sql.NullString{},
			"see",
			sql.NullString{},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := cd.EnsurePrivateForumTopicSeeGrant(userID); err != nil {
		t.Fatalf("EnsurePrivateForumTopicSeeGrant: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEnsurePrivateForumTopicSeeGrantReusesExistingGrant(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cd := NewTestCoreData(t, db.New(conn))
	const userID int32 = 23

	mock.ExpectQuery("WITH role_ids AS \\(").
		WithArgs(
			userID,
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"see",
			sql.NullInt32{},
			sql.NullInt32{Int32: userID, Valid: true},
		).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	if err := cd.EnsurePrivateForumTopicSeeGrant(userID); err != nil {
		t.Fatalf("EnsurePrivateForumTopicSeeGrant: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
