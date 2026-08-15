package common

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestCreatePrivateTopicWithAccessAuthorizesBeforeWrites(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cd := NewTestCoreData(t, db.New(conn))
	cd.UserID = 1

	mock.ExpectQuery("WITH role_ids AS \\(").
		WithArgs(int32(1), "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}))

	_, err = cd.CreatePrivateTopicWithAccess(CreatePrivateTopicParams{
		CreatorID: 1,
		Participants: []PrivateTopicParticipant{
			{ID: 2, Username: "participant"},
		},
		Title: "Private topic",
	})
	if err == nil {
		t.Fatal("CreatePrivateTopicWithAccess unexpectedly succeeded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreatePrivateTopicWithAccessCreatesMissingGlobalSeeGrantTransactionally(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cd := NewTestCoreData(t, db.New(conn))
	cd.UserID = 1

	// userID identifies the invited participant whose access is established.
	const userID int32 = 23
	// topicID is the database ID returned for the newly created private topic.
	const topicID int64 = 42

	mock.ExpectQuery("WITH role_ids AS \\(").
		WithArgs(int32(1), "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT idusers FROM users WHERE idusers = \\? FOR UPDATE").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"idusers"}).AddRow(userID))
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
	mock.ExpectExec("INSERT INTO forumtopic").
		WithArgs(
			PrivateForumCategoryID,
			sql.NullInt32{},
			sql.NullString{String: "Private topic", Valid: true},
			sql.NullString{String: "", Valid: true},
			"private",
			"privateforum",
			sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
			sql.NullInt32{Int32: 1, Valid: true},
			int32(1),
		).
		WillReturnResult(sqlmock.NewResult(topicID, 1))
	for _, action := range []string{"see", "view", "post", "reply"} {
		mock.ExpectExec("INSERT INTO grants").
			WithArgs(
				sql.NullInt32{Int32: userID, Valid: true},
				sql.NullInt32{},
				"privateforum",
				sql.NullString{String: "topic", Valid: true},
				"allow",
				sql.NullInt32{Int32: int32(topicID), Valid: true},
				sql.NullString{},
				action,
				sql.NullString{},
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	gotTopicID, err := cd.CreatePrivateTopicWithAccess(CreatePrivateTopicParams{
		CreatorID: 1,
		Participants: []PrivateTopicParticipant{
			{ID: userID, Username: "participant"},
		},
		Title: "Private topic",
	})
	if err != nil {
		t.Fatalf("CreatePrivateTopicWithAccess: %v", err)
	}
	if gotTopicID != int32(topicID) {
		t.Fatalf("CreatePrivateTopicWithAccess topic ID = %d, want %d", gotTopicID, topicID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
