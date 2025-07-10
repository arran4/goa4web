package forum

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	db "github.com/arran4/goa4web/internal/db"
)

func TestCanCreateThreadAllowed(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "forumtopic_idforumtopic", "viewlevel", "replylevel", "newthreadlevel", "seelevel", "invitelevel", "readlevel", "modlevel", "adminlevel",
		}).AddRow(1, 1, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{}))

	mock.ExpectQuery("SELECT utl.*").
		WithArgs(int32(2), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers", "forumtopic_idforumtopic", "level", "invitemax", "expires_at"}).
			AddRow(2, 1, sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{}, sql.NullTime{}))

	ok, err := canCreateThread(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("canCreateThread: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestCanCreateThreadDenied(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "forumtopic_idforumtopic", "viewlevel", "replylevel", "newthreadlevel", "seelevel", "invitelevel", "readlevel", "modlevel", "adminlevel",
		}).AddRow(1, 1, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{Int32: 5, Valid: true}, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt32{}))

	mock.ExpectQuery("SELECT utl.*").
		WithArgs(int32(2), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers", "forumtopic_idforumtopic", "level", "invitemax", "expires_at"}).
			AddRow(2, 1, sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{}, sql.NullTime{}))

	ok, err := canCreateThread(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("canCreateThread: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
