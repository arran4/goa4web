package common

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestCreatePrivateTopicUsesProvidedUsernames(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	queries := db.New(conn)
	cd := NewTestCoreData(t, queries)
	cd.UserID = 1

	mock.ExpectQuery("WITH role_ids AS \\(").WithArgs(int32(1), "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, false, sql.NullInt32{Int32: 1, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("WITH role_ids AS \\(").WithArgs(int32(2), "privateforum", sql.NullString{String: "topic", Valid: true}, "see", sql.NullInt32{}, false, sql.NullInt32{Int32: 2, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	topicID := int64(42)
	expectedTitle := "Private chat with creator, participant"
	mock.ExpectExec("INSERT INTO forumtopic").
		WithArgs(
			PrivateForumCategoryID,
			sql.NullInt32{Int32: 1, Valid: true},
			sql.NullString{String: expectedTitle, Valid: true},
			sql.NullString{String: expectedTitle, Valid: true},
			"private",
			"privateforum",
			sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
			sql.NullInt32{Int32: 1, Valid: true},
			int32(1),
		).WillReturnResult(sqlmock.NewResult(topicID, 1))

	for _, uid := range []int32{1, 2} {
		for _, act := range []string{"see", "view", "post", "reply"} {
			mock.ExpectExec("INSERT INTO grants").
				WithArgs(
					sql.NullInt32{Int32: uid, Valid: true},
					sql.NullInt32{},
					"privateforum",
					sql.NullString{String: "topic", Valid: true},
					"allow",
					sql.NullInt32{Int32: int32(topicID), Valid: true},
					sql.NullString{},
					act,
					sql.NullString{},
				).WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectExec("INSERT INTO subscriptions").
			WithArgs(
				uid,
				"create thread:/forum/topic/42/*",
				"internal",
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	tid, err := cd.CreatePrivateTopic(CreatePrivateTopicParams{
		CreatorID: 1,
		Participants: []PrivateTopicParticipant{
			{ID: 1, Username: "creator"},
			{ID: 2, Username: "participant"},
		},
	})
	if err != nil {
		t.Fatalf("CreatePrivateTopic: %v", err)
	}
	if tid != int32(topicID) {
		t.Fatalf("CreatePrivateTopic topic id = %d, want %d", tid, topicID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreatePrivateTopicBuildsUsernamesWhenMissing(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	queries := db.New(conn)
	cd := NewTestCoreData(t, queries)
	cd.UserID = 1

	mock.ExpectQuery("WITH role_ids AS \\(").WithArgs(int32(1), "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, false, sql.NullInt32{Int32: 1, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("WITH role_ids AS \\(").WithArgs(int32(2), "privateforum", sql.NullString{String: "topic", Valid: true}, "see", sql.NullInt32{}, false, sql.NullInt32{Int32: 2, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).
			AddRow(1, nil, "creator", nil))
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at").
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).
			AddRow(2, nil, "participant", nil))

	topicID := int64(7)
	expectedTitle := "Private chat with creator, participant"
	mock.ExpectExec("INSERT INTO forumtopic").
		WithArgs(
			PrivateForumCategoryID,
			sql.NullInt32{Int32: 1, Valid: true},
			sql.NullString{String: expectedTitle, Valid: true},
			sql.NullString{String: expectedTitle, Valid: true},
			"private",
			"privateforum",
			sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
			sql.NullInt32{Int32: 1, Valid: true},
			int32(1),
		).WillReturnResult(sqlmock.NewResult(topicID, 1))

	for _, uid := range []int32{1, 2} {
		for _, act := range []string{"see", "view", "post", "reply"} {
			mock.ExpectExec("INSERT INTO grants").
				WithArgs(
					sql.NullInt32{Int32: uid, Valid: true},
					sql.NullInt32{},
					"privateforum",
					sql.NullString{String: "topic", Valid: true},
					"allow",
					sql.NullInt32{Int32: int32(topicID), Valid: true},
					sql.NullString{},
					act,
					sql.NullString{},
				).WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectExec("INSERT INTO subscriptions").
			WithArgs(
				uid,
				"create thread:/forum/topic/7/*",
				"internal",
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	tid, err := cd.CreatePrivateTopic(CreatePrivateTopicParams{
		CreatorID: 1,
		Participants: []PrivateTopicParticipant{
			{ID: 1},
			{ID: 2},
		},
	})
	if err != nil {
		t.Fatalf("CreatePrivateTopic: %v", err)
	}
	if tid != int32(topicID) {
		t.Fatalf("CreatePrivateTopic topic id = %d, want %d", tid, topicID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreatePrivateTopicRejectsIneligibleParticipants(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	queries := db.New(conn)
	cd := NewTestCoreData(t, queries)
	cd.UserID = 1

	// Creator check succeeds
	mock.ExpectQuery("WITH role_ids AS \\(").
		WithArgs(int32(1), "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, false, sql.NullInt32{Int32: 1, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// Participant check fails (ineligible / no rows)
	mock.ExpectQuery("WITH role_ids AS \\(").
		WithArgs(int32(2), "privateforum", sql.NullString{String: "topic", Valid: true}, "see", sql.NullInt32{}, false, sql.NullInt32{Int32: 2, Valid: true}, false).
		WillReturnError(sql.ErrNoRows)

	tid, err := cd.CreatePrivateTopic(CreatePrivateTopicParams{
		CreatorID: 1,
		Participants: []PrivateTopicParticipant{
			{ID: 1, Username: "creator"},
			{ID: 2, Username: "ineligible_user"},
		},
	})
	if err == nil {
		t.Fatalf("expected CreatePrivateTopic to fail for ineligible participant, got topic %d", tid)
	}

	var errInvalid *ErrInvalidParticipants
	if !errors.As(err, &errInvalid) {
		t.Fatalf("expected ErrInvalidParticipants, got: %v", err)
	}
	if len(errInvalid.Usernames) != 1 || errInvalid.Usernames[0] != "ineligible_user" {
		t.Errorf("unexpected invalid usernames: %v", errInvalid.Usernames)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
