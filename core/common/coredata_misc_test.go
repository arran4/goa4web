package common

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"

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

	topicID := int64(42)
	expectedTitle := "Private chat with creator, participant"
	mock.ExpectExec("INSERT INTO forumtopic").
		WithArgs(
			PrivateForumCategoryID,
			sql.NullInt32{},
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
			sql.NullInt32{},
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

func TestMapLinkURL(t *testing.T) {
	cd := &CoreData{
		Config: &config.RuntimeConfig{
			BaseURL: "http://example.com",
		},
		LinkSignKey: "test-link-key",
	}

	tests := []struct {
		name string
		tag  string
		val  string
		want string // we'll just check if it was signed or not by looking at prefixes or identical values
	}{
		{"not an anchor tag", "img", "http://external.com/img.png", "http://external.com/img.png"},
		{"relative url", "a", "/internal/path", "/internal/path"},
		{"allowed host", "a", "http://example.com/path", "http://example.com/path"},
		{"allowed host case insensitive", "a", "HTTP://eXaMpLe.CoM/path", "HTTP://eXaMpLe.CoM/path"},
		{"disallowed host", "a", "http://external.com/path", "signed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cd.MapLinkURL(tt.tag, tt.val)

			if tt.want == "signed" {
				if got == tt.val || !strings.Contains(got, "/goto?u=") {
					t.Errorf("MapLinkURL(%q, %q) = %q, want signed URL", tt.tag, tt.val, got)
				}
			} else {
				if got != tt.want {
					t.Errorf("MapLinkURL(%q, %q) = %q, want %q", tt.tag, tt.val, got, tt.want)
				}
			}
		})
	}
}
