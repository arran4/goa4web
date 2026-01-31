package common

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestGetPrivateTopicDisplayTitle(t *testing.T) {
	tests := []struct {
		name              string
		userID            int32
		originalTitle     string
		participants      []*db.AdminListPrivateTopicParticipantsByTopicIDRow
		expectedTitle     string
		expectQuery       bool
	}{
		{
			name:          "Logged In (User 1) - Two Participants",
			userID:        1,
			originalTitle: "Private chat with Hidden",
			participants: []*db.AdminListPrivateTopicParticipantsByTopicIDRow{
				{Idusers: 1, Username: sql.NullString{String: "Me", Valid: true}},
				{Idusers: 2, Username: sql.NullString{String: "Alice", Valid: true}},
			},
			expectedTitle: "Alice",
			expectQuery:   true,
		},
		{
			name:          "Logged In (User 1) - Only Me",
			userID:        1,
			originalTitle: "Private chat with Hidden",
			participants: []*db.AdminListPrivateTopicParticipantsByTopicIDRow{
				{Idusers: 1, Username: sql.NullString{String: "Me", Valid: true}},
			},
			expectedTitle: "Me", // Fallback to myself
			expectQuery:   true,
		},
		{
			name:          "Logged In (User 1) - Others Only",
			userID:        1,
			originalTitle: "Private chat with Hidden",
			participants: []*db.AdminListPrivateTopicParticipantsByTopicIDRow{
				{Idusers: 2, Username: sql.NullString{String: "Alice", Valid: true}},
				{Idusers: 3, Username: sql.NullString{String: "Bob", Valid: true}},
			},
			expectedTitle: "Alice, Bob",
			expectQuery:   true,
		},
		{
			name:          "Guest (User 0) - Two Participants",
			userID:        0,
			originalTitle: "Private chat with Hidden",
			participants: []*db.AdminListPrivateTopicParticipantsByTopicIDRow{
				{Idusers: 1, Username: sql.NullString{String: "Me", Valid: true}},
				{Idusers: 2, Username: sql.NullString{String: "Alice", Valid: true}},
			},
			expectedTitle: "Me, Alice", // Should show everyone
			expectQuery:   true,
		},
		{
			name:          "Custom Title (No Prefix)",
			userID:        1,
			originalTitle: "My Project Discussion",
			participants:  nil,
			expectedTitle: "My Project Discussion",
			expectQuery:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer conn.Close()

			if tt.expectQuery {
				rows := sqlmock.NewRows([]string{"idusers", "username"})
				for _, p := range tt.participants {
					rows.AddRow(p.Idusers, p.Username)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, u.username")).
					WithArgs(sql.NullInt32{Int32: 123, Valid: true}). // Assuming topicID 123
					WillReturnRows(rows)
			}

			queries := db.New(conn)
			cd := NewTestCoreData(t, queries)
			cd.UserID = tt.userID

			title := cd.GetPrivateTopicDisplayTitle(123, tt.originalTitle)

			if title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, title)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations: %v", err)
			}
		})
	}
}
