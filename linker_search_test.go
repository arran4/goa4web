package main

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLinkerApproveAddsToSearch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := New(db)

	// Approve item from queue id 1
	mock.ExpectExec("INSERT INTO linker").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{"idlinker", "language_idlanguage", "users_idusers", "linkercategory_idlinkerCategory", "forumthread_idforumthread", "title", "url", "description", "listed", "username", "title_2"}).
		AddRow(1, 1, 1, 1, 0, "Foo", "http://foo", "Bar", time.Now(), "u", "c")
	mock.ExpectQuery("SELECT l.idlinker").WithArgs(int32(1)).WillReturnRows(rows)

	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("foo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT IGNORE INTO linkerSearch").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").WithArgs("bar").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT IGNORE INTO linkerSearch").WithArgs(int32(1), int32(2)).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/admin/queue?qid=1", nil)
	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	linkerAdminQueueApproveActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
