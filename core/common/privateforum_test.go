package common

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestCoreData_PrivateForumTopics(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	cd := NewTestCoreData(t, q)
	cd.UserID = 1

	rows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "LastPosterUsername"}).
		AddRow(1, 1, 0, 1, "Test Topic", "Test Description", 1, 1, time.Now(), "private", "testuser")
	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(int64(1), "privateforum", "topic", "see", nil, int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT t.idforumtopic, t.lastposter, t.forumcategory_idforumcategory, t.language_id, t.title, t.description, t.threads, t.comments, t.lastaddition, t.handler, lu.username AS LastPosterUsername").
		WithArgs(sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(rows)

	participantRows := sqlmock.NewRows([]string{"idusers", "username"}).
		AddRow(2, "participant1")
	mock.ExpectQuery("SELECT u.idusers, u.username").
		WithArgs(sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(participantRows)
	mock.ExpectQuery("SELECT item, item_id, label").
		WithArgs("thread", int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))

	topics, err := cd.PrivateForumTopics()
	if err != nil {
		t.Fatalf("PrivateForumTopics() error = %v", err)
	}

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	if topics[0].Idforumtopic != 1 {
		t.Errorf("expected topic id 1, got %d", topics[0].Idforumtopic)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
