package common

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestCreatePrivateTopicUsesProvidedUsernames(t *testing.T) {
	queries := testutil.NewPrivateTopicQuerier(t)
	queries.AllowGrants()
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), WithUserRoles([]string{"user"}))
	cd.UserID = 1

	topicID := int64(42)
	expectedTitle := "Private chat with creator, participant"
	queries.CreatedTopicID = topicID

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
	if queries.CreatedTopic.Title.String != expectedTitle {
		t.Fatalf("CreateForumTopicForPoster title %q, want %q", queries.CreatedTopic.Title.String, expectedTitle)
	}

}

func TestCreatePrivateTopicBuildsUsernamesWhenMissing(t *testing.T) {
	queries := testutil.NewPrivateTopicQuerier(t)
	queries.AllowGrants()
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), WithUserRoles([]string{"user"}))
	cd.UserID = 1
	queries.Users[1] = &db.SystemGetUserByIDRow{Idusers: 1, Username: sql.NullString{String: "creator", Valid: true}}
	queries.Users[2] = &db.SystemGetUserByIDRow{Idusers: 2, Username: sql.NullString{String: "participant", Valid: true}}

	topicID := int64(7)
	expectedTitle := "Private chat with creator, participant"
	queries.CreatedTopicID = topicID

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
	if queries.CreatedTopic.Title.String != expectedTitle {
		t.Fatalf("CreateForumTopicForPoster title %q, want %q", queries.CreatedTopic.Title.String, expectedTitle)
	}

}
