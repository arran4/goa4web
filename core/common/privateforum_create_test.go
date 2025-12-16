package common

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestCreatePrivateTopic(t *testing.T) {
	var topicID int64 = 1
	var threadID int64 = 1
	var commentID int64 = 1
	var grantID int64
	var grantCount int
	queries := &db.QuerierProxier{
		Querier: nil,
		OverwrittenSystemCheckGrant: func(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
			return 1, nil
		},
		OverwrittenGetPermissionsByUserID: func(ctx context.Context, usersIdusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
			return []*db.GetPermissionsByUserIDRow{}, nil
		},
		OverwrittenSystemCheckRoleGrant: func(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
			return 1, nil
		},
		OverwrittenSystemGetUserByID: func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Username: sql.NullString{String: "testuser", Valid: true},
			}, nil
		},
		OverwrittenListPrivateTopicParticipantsByTopicIDForUser: func(ctx context.Context, arg db.ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*db.ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
			return []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{}, nil
		},
		OverwrittenCreateForumTopicForPoster: func(ctx context.Context, arg db.CreateForumTopicForPosterParams) (int64, error) {
			return topicID, nil
		},
		OverwrittenCreateForumThreadForPoster: func(ctx context.Context, arg db.CreateForumThreadForPosterParams) (int64, error) {
			return threadID, nil
		},
		OverwrittenCreateCommentInSectionForCommenter: func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return commentID, nil
		},
		OverwrittenSystemCreateGrant: func(ctx context.Context, arg db.SystemCreateGrantParams) (int64, error) {
			grantID++
			grantCount++
			return grantID, nil
		},
	}
	cd := &CoreData{
		queries: queries,
		ctx:     context.Background(),
		UserID:  1,
	}

	params := CreatePrivateTopicParams{
		CreatorID:      1,
		ParticipantIDs: []int32{1, 2},
		Title:          "Test Topic",
		Description:    "Test Description",
	}

	actualTopicID, err := cd.CreatePrivateTopic(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, int32(topicID), actualTopicID, fmt.Sprintf("expected topic ID %v, got %d", topicID, actualTopicID))
	assert.Equal(t, 10, grantCount, "expected grant count to be 10")

}
