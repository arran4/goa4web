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
	// This ensures we are not calling it or using it.
	q.ListContentPublicLabelsReturns = []*db.ListContentPublicLabelsRow{
		{
			Item:   "thread",
			ItemID: 1,
			Label:  "Misapplied",
		},
	}

	// Stub GetPrivateTopicThreadsAndLabels to return a simple thread with no special labels (so Read and NotNew effectively if we assume default... wait)
	// Default:
	// Unread: If "unread" label missing -> Unread.
	// New: If "new" label missing -> New.
	// But we want to test "Bug Fix", i.e. no random labels.
	// If we return empty list, no labels.
	// If we return a thread with "unread" inverted=true and "new" inverted=true, we should have NO labels.
	q.GetPrivateTopicThreadsAndLabelsReturns = []*db.GetPrivateTopicThreadsAndLabelsRow{
		{
			Idforumthread: 100,
			AuthorID:      2, // Other user
			Label:         sql.NullString{String: "unread", Valid: true},
			Invert:        sql.NullBool{Bool: true, Valid: true},
		},
		{
			Idforumthread: 100,
			AuthorID:      2,
			Label:         sql.NullString{String: "new", Valid: true},
			Invert:        sql.NullBool{Bool: true, Valid: true},
		},
	}

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

	// Stub GetPrivateTopicThreadsAndLabels
	// Thread 101: Author=2 (other). No labels. -> Should be Unread and New.
	q.GetPrivateTopicThreadsAndLabelsReturns = []*db.GetPrivateTopicThreadsAndLabelsRow{
		{
			Idforumthread: 101,
			AuthorID:      2,
			Label:         sql.NullString{Valid: false}, // No labels
			Invert:        sql.NullBool{Valid: false},
		},
	}

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

func TestCoreData_PrivateForumTopics_OwnThreadNotNew(t *testing.T) {
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("privateforum", "topic", "see"))
	cd := NewTestCoreData(t, q)
	cd.UserID = 1

	q.ListPrivateTopicsByUserIDReturns = []*db.ListPrivateTopicsByUserIDRow{
		{
			Idforumtopic: 1,
			Handler:      "private",
		},
	}
	q.ListPrivateTopicParticipantsByTopicIDForUserReturns = []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{}

	// Thread 102: Author=1 (Me). No labels. -> Should be Unread (if not read) but NOT New (because I wrote it).
	q.GetPrivateTopicThreadsAndLabelsReturns = []*db.GetPrivateTopicThreadsAndLabelsRow{
		{
			Idforumthread: 102,
			AuthorID:      1, // Me
			Label:         sql.NullString{Valid: false},
			Invert:        sql.NullBool{Valid: false},
		},
	}

	topics, _ := cd.PrivateForumTopics()
	foundNew := false
	for _, l := range topics[0].Labels {
		if l.Name == "new" {
			foundNew = true
		}
	}
	if foundNew {
		t.Errorf("Did not expect 'new' label for own thread.")
	}
}
