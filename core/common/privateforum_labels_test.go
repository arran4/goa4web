package common

import (
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCoreData_PrivateForumTopics_LabelsBug(t *testing.T) {
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("privateforum", "topic", "see"))
	cd := NewTestCoreData(t, q)
	cd.UserID = 1

	// Setup a private topic with ID 1
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

	// Setup ListContentPublicLabels to return a "Misapplied" label for item="thread" and itemID=1
	// Even though we fixed the code, the mock setup for this remains to ensure we are NOT calling it anymore
	// or ignoring it.
	q.ListContentPublicLabelsReturns = []*db.ListContentPublicLabelsRow{
		{
			Item:   "thread",
			ItemID: 1,
			Label:  "Misapplied",
		},
	}

	// Stub GetPrivateTopicReadStatus to return no unread/new
	q.GetPrivateTopicReadStatusReturns = &db.GetPrivateTopicReadStatusRow{HasUnread: false, HasNew: false}

	topics, err := cd.PrivateForumTopics()
	if err != nil {
		t.Fatalf("PrivateForumTopics() error = %v", err)
	}

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	// Check if the "Misapplied" label is present
	found := false
	for _, l := range topics[0].Labels {
		if l.Name == "Misapplied" {
			found = true
			break
		}
	}

	if found {
		t.Errorf("Did NOT expect 'Misapplied' label to be present (bug fixed), but it WAS found.")
	} else {
		t.Logf("Confirmed: 'Misapplied' label was NOT found on the topic. Bug fixed.")
	}
}

func TestCoreData_PrivateForumTopics_UnreadNew(t *testing.T) {
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("privateforum", "topic", "see"))
	cd := NewTestCoreData(t, q)
	cd.UserID = 1

	q.ListPrivateTopicsByUserIDReturns = []*db.ListPrivateTopicsByUserIDRow{
		{
			Idforumtopic:                 1,
			Lastposter:                   1,
			ForumcategoryIdforumcategory: 0,
			Title:                        sql.NullString{String: "Test Topic", Valid: true},
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

	// Stub GetPrivateTopicReadStatus to return HasUnread=true, HasNew=true
	q.GetPrivateTopicReadStatusReturns = &db.GetPrivateTopicReadStatusRow{HasUnread: true, HasNew: true}

	topics, err := cd.PrivateForumTopics()
	if err != nil {
		t.Fatalf("PrivateForumTopics() error = %v", err)
	}

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	foundUnread := false
	foundNew := false
	for _, l := range topics[0].Labels {
		if l.Name == "unread" && l.Type == "private" {
			foundUnread = true
		}
		if l.Name == "new" && l.Type == "private" {
			foundNew = true
		}
	}

	if !foundUnread {
		t.Errorf("Expected 'unread' label, but not found.")
	}
	if !foundNew {
		t.Errorf("Expected 'new' label, but not found.")
	}
}
