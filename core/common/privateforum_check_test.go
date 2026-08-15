package common

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type privateForumCheckQuerier struct {
	db.Querier
	grants  []*db.AdminListAllPrivateForumGrantsRow
	threads []*db.AdminListAllPrivateForumThreadsRow
}

func (q *privateForumCheckQuerier) AdminListAllPrivateForumGrants(context.Context) ([]*db.AdminListAllPrivateForumGrantsRow, error) {
	return q.grants, nil
}

func (q *privateForumCheckQuerier) AdminListAllPrivateForumThreads(context.Context) ([]*db.AdminListAllPrivateForumThreadsRow, error) {
	return q.threads, nil
}

func TestCheckPrivateForumInconsistenciesPreservesFineGrainedThreadGrants(t *testing.T) {
	const (
		userID           int32 = 7
		topicID          int32 = 11
		allowedThreadID  int32 = 22
		excludedThreadID int32 = 23
	)
	queries := &privateForumCheckQuerier{
		grants: []*db.AdminListAllPrivateForumGrantsRow{
			{
				ID:       1,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionView.String(),
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
			{
				ID:       2,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionView.String(),
				ItemID:   sql.NullInt32{Int32: allowedThreadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
			{
				ID:       3,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionReply.String(),
				ItemID:   sql.NullInt32{Int32: allowedThreadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
		},
		threads: []*db.AdminListAllPrivateForumThreadsRow{
			{Idforumthread: allowedThreadID, Idforumtopic: topicID},
			{Idforumthread: excludedThreadID, Idforumtopic: topicID},
		},
	}
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	inconsistencies, err := cd.CheckAndFixPrivateForumInconsistencies(context.Background(), nil, true)
	if err != nil {
		t.Fatalf("CheckAndFixPrivateForumInconsistencies: %v", err)
	}
	if len(inconsistencies) != 0 {
		t.Fatalf("fine-grained thread grants reported as inconsistent: %+v", inconsistencies)
	}
}
