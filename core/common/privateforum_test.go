package common

import (
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCoreData_PrivateForumTopics(t *testing.T) {
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("privateforum", "topic", "see"))
	cd := NewTestCoreData(t, q)
	cd.UserID = 1

	q.ListPrivateTopicsByUserIDReturns = []*db.ListPrivateTopicsByUserIDRow{
		{
			Idforumtopic:                 1,
			Lastposter:                   1,
			ForumcategoryIdforumcategory: 0,
			Title:                        sql.NullString{String: "Test Topic", Valid: true},
			Description:                  sql.NullString{String: "Test Description", Valid: true},
			Threads:                      sql.NullInt32{Int32: 1, Valid: true},
			Comments:                     sql.NullInt32{Int32: 1, Valid: true},
			Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			Handler:                      "private",
			Lastposterusername:           sql.NullString{String: "testuser", Valid: true},
		},
	}
	q.ListPrivateTopicParticipantsByTopicIDForUserReturns = []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
		{Idusers: 2, Username: sql.NullString{String: "participant1", Valid: true}},
	}

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
}
