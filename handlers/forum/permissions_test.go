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
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic, r.forumtopic_idforumtopic, r.view_role_id, r.reply_role_id, r.newthread_role_id, r.see_role_id, r.invite_role_id, r.read_role_id, r.mod_role_id, r.admin_role_id FROM forumtopic t LEFT JOIN topic_permissions r ON t.idforumtopic = r.forumtopic_idforumtopic WHERE idforumtopic = ?")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "forumtopic_idforumtopic", "view_role_id", "reply_role_id", "newthread_role_id", "see_role_id", "invite_role_id", "read_role_id", "mod_role_id", "admin_role_id"}).AddRow(1, 1, nil, nil, 2, nil, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT utl.users_idusers, utl.forumtopic_idforumtopic, utl.role_id, utl.invitemax, utl.expires_at FROM user_topic_permissions utl WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?")).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers", "forumtopic_idforumtopic", "role_id", "invitemax", "expires_at"}).AddRow(2, 1, 3, nil, nil))

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
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic, r.forumtopic_idforumtopic, r.view_role_id, r.reply_role_id, r.newthread_role_id, r.see_role_id, r.invite_role_id, r.read_role_id, r.mod_role_id, r.admin_role_id FROM forumtopic t LEFT JOIN topic_permissions r ON t.idforumtopic = r.forumtopic_idforumtopic WHERE idforumtopic = ?")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "forumtopic_idforumtopic", "view_role_id", "reply_role_id", "newthread_role_id", "see_role_id", "invite_role_id", "read_role_id", "mod_role_id", "admin_role_id"}).AddRow(1, 1, nil, nil, 3, nil, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT utl.users_idusers, utl.forumtopic_idforumtopic, utl.role_id, utl.invitemax, utl.expires_at FROM user_topic_permissions utl WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?")).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers", "forumtopic_idforumtopic", "role_id", "invitemax", "expires_at"}).AddRow(2, 1, 1, nil, nil))

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
