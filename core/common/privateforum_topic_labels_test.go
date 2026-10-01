package common

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCoreData_PrivateForumTopics_ShowsTopicLabels(t *testing.T) {
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
			Threads:                      sql.NullInt32{Int32: 0, Valid: true},
			Comments:                     sql.NullInt32{Int32: 0, Valid: true},
			Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			Handler:                      "private",
			Lastposterusername:           sql.NullString{String: "testuser", Valid: true},
		},
	}

	// No derived thread labels (empty)
	q.GetPrivateTopicThreadsAndLabelsForUserReturns = []*db.GetPrivateTopicThreadsAndLabelsForUserRow{}

	ctx := context.TODO()

	// Public Label
	if err := q.AddContentPublicLabel(ctx, db.AddContentPublicLabelParams{
		Item:   "topic",
		ItemID: 1,
		Label:  "PublicTag",
	}); err != nil {
		t.Fatalf("failed to add public label: %v", err)
	}

	// Private Label
	if err := q.AddContentPrivateLabel(ctx, db.AddContentPrivateLabelParams{
		Item:   "topic",
		ItemID: 1,
		UserID: 1,
		Label:  "PrivateTag",
		Invert: false,
	}); err != nil {
		t.Fatalf("failed to add private label: %v", err)
	}

	topics := testhelpers.Must(cd.PrivateForumTopics())

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	foundPublic := false
	foundPrivate := false

	for _, l := range topics[0].Labels {
		if l.Name == "PublicTag" && l.Type == "public" {
			foundPublic = true
		}
		if l.Name == "PrivateTag" && l.Type == "private" {
			foundPrivate = true
		}
	}

	if !foundPublic {
		t.Errorf("Expected 'PublicTag' label of type 'public', but not found.")
	}
	if !foundPrivate {
		t.Errorf("Expected 'PrivateTag' label of type 'private', but not found.")
	}
}

func TestAddTopicPublicLabelAction_AuthorizationInvariant(t *testing.T) {
	// User 2 has unscoped privateforum topic label capability
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("privateforum", "topic", "label"))
	cd := NewTestCoreData(t, q)
	cd.UserID = 2

	ctx := context.Background()
	topicID := int32(42)

	// User 2 is NOT a participant in topic 42 (GetForumTopicByIdForUser returns ErrNoRows)
	q.GetForumTopicByIdForUserErr = sql.ErrNoRows

	err := cd.AddTopicPublicLabelAction(ctx, TopicLabelParams{
		ActorID: 2,
		TopicID: topicID,
		Label:   "unauthorized-tag",
	})
	if err == nil {
		t.Fatal("expected error for user who has label capability but cannot access target topic, got nil")
	}

	// Verify state remained unchanged (no public labels added)
	pubLabels, err := q.ListContentPublicLabels(ctx, db.ListContentPublicLabelsParams{
		Item:   "topic",
		ItemID: topicID,
	})
	if err != nil {
		t.Fatalf("failed to query public labels: %v", err)
	}
	if len(pubLabels) > 0 {
		t.Errorf("expected 0 public labels, got %d", len(pubLabels))
	}

	// Now make User 2 a participant in topic 42 (GetForumTopicByIdForUser returns topic)
	q.GetForumTopicByIdForUserErr = nil
	q.GetForumTopicByIdForUserReturns = &db.GetForumTopicByIdForUserRow{
		Idforumtopic: topicID,
		Handler:      "private",
	}

	err = cd.AddTopicPublicLabelAction(ctx, TopicLabelParams{
		ActorID: 2,
		TopicID: topicID,
		Label:   "authorized-tag",
	})
	if err != nil {
		t.Fatalf("expected successful label add for authorized participant with capability, got: %v", err)
	}
}
