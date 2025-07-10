package forum

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	db "github.com/arran4/goa4web/internal/db"
)

func TestUserCanCreateThread_Allowed(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic, r.forumtopic_idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, r.modlevel, r.adminlevel FROM forumtopic t LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic WHERE idforumtopic = ?")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "forumtopic_idforumtopic", "viewlevel", "replylevel", "newthreadlevel", "seelevel", "invitelevel", "readlevel", "modlevel", "adminlevel"}).AddRow(1, 1, nil, nil, 2, nil, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT utl.users_idusers, utl.forumtopic_idforumtopic, utl.level, utl.invitemax, utl.expires_at FROM userstopiclevel utl WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?")).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers", "forumtopic_idforumtopic", "level", "invitemax", "expires_at"}).AddRow(2, 1, 3, nil, nil))

	ok, err := UserCanCreateThread(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateThread: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserCanCreateThread_Denied(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic, r.forumtopic_idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, r.modlevel, r.adminlevel FROM forumtopic t LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic WHERE idforumtopic = ?")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "forumtopic_idforumtopic", "viewlevel", "replylevel", "newthreadlevel", "seelevel", "invitelevel", "readlevel", "modlevel", "adminlevel"}).AddRow(1, 1, nil, nil, 3, nil, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT utl.users_idusers, utl.forumtopic_idforumtopic, utl.level, utl.invitemax, utl.expires_at FROM userstopiclevel utl WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?")).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers", "forumtopic_idforumtopic", "level", "invitemax", "expires_at"}).AddRow(2, 1, 1, nil, nil))

	ok, err := UserCanCreateThread(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateThread: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
