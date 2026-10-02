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
	grants        []*db.AdminListAllPrivateForumGrantsRow
	threads       []*db.AdminListAllPrivateForumThreadsRow
	deletedGrants []int32
}

func (q *privateForumCheckQuerier) AdminListAllPrivateForumGrants(context.Context) ([]*db.AdminListAllPrivateForumGrantsRow, error) {
	return q.grants, nil
}

func (q *privateForumCheckQuerier) AdminListAllPrivateForumThreads(context.Context) ([]*db.AdminListAllPrivateForumThreadsRow, error) {
	return q.threads, nil
}

func (q *privateForumCheckQuerier) AdminDeleteGrant(_ context.Context, id int32) error {
	q.deletedGrants = append(q.deletedGrants, id)
	return nil
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
			{
				ID:       4,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionEdit.String(),
				ItemID:   sql.NullInt32{Int32: allowedThreadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
			{
				ID:       5,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionEditAny.String(),
				ItemID:   sql.NullInt32{Int32: allowedThreadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
			{
				ID:       6,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   "append",
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

func TestCheckPrivateForumInconsistenciesCapabilityGrants(t *testing.T) {
	const (
		userID  int32 = 7
		topicID int32 = 11
	)
	queries := &privateForumCheckQuerier{
		grants: []*db.AdminListAllPrivateForumGrantsRow{
			// Unscoped capability grants on privateforum/topic (role-based and user-based)
			{
				ID:       1,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionEdit.String(),
				ItemID:   sql.NullInt32{Valid: false},
				RoleName: sql.NullString{String: "admin", Valid: true},
			},
			{
				ID:       2,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionLabel.String(),
				ItemID:   sql.NullInt32{Valid: false},
				RoleName: sql.NullString{String: "user", Valid: true},
			},
			{
				ID:       3,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionCreate.String(),
				ItemID:   sql.NullInt32{Valid: false},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "alice", Valid: true},
			},
			{
				ID:       4,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionSee.String(),
				ItemID:   sql.NullInt32{Valid: false},
				RoleName: sql.NullString{String: "user", Valid: true},
			},
			{
				ID:       5,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionView.String(),
				ItemID:   sql.NullInt32{Valid: false},
				RoleName: sql.NullString{String: "user", Valid: true},
			},
			// Scoped edit grant on specific topic
			{
				ID:       6,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionEdit.String(),
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "alice", Valid: true},
			},
			// Scoped label grant on specific topic
			{
				ID:       7,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionLabel.String(),
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "alice", Valid: true},
			},
			// Corrupt grant: thread grant without item_id
			{
				ID:       8,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionView.String(),
				ItemID:   sql.NullInt32{Valid: false},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "alice", Valid: true},
			},
			// Corrupt grant: 'anyone' access
			{
				ID:       9,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionView.String(),
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				UserID:   sql.NullInt32{Valid: false},
				RoleName: sql.NullString{Valid: false},
			},
		},
		threads: []*db.AdminListAllPrivateForumThreadsRow{
			{Idforumthread: 100, Idforumtopic: topicID},
		},
	}
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	inconsistencies, err := cd.CheckAndFixPrivateForumInconsistencies(context.Background(), nil, true)
	if err != nil {
		t.Fatalf("CheckAndFixPrivateForumInconsistencies: %v", err)
	}

	// Grants 1-7 must NOT be flagged as inconsistent.
	// Grant 8 (thread grant without item_id) and 9 (anyone grant) must be flagged.
	if len(inconsistencies) != 2 {
		t.Fatalf("expected 2 inconsistencies, got %d: %+v", len(inconsistencies), inconsistencies)
	}

	foundGrant8 := false
	foundGrant9 := false
	for _, inc := range inconsistencies {
		if inc.GrantID == 8 {
			foundGrant8 = true
			if inc.Issue != "Grant has no item_id (access to all topics/threads)" {
				t.Errorf("grant 8 issue = %q", inc.Issue)
			}
		}
		if inc.GrantID == 9 {
			foundGrant9 = true
			if inc.Issue != "Grant allows 'anyone' access (no role, no user)" {
				t.Errorf("grant 9 issue = %q", inc.Issue)
			}
		}
	}
	if !foundGrant8 || !foundGrant9 {
		t.Errorf("expected grants 8 and 9 flagged, got foundGrant8=%v, foundGrant9=%v", foundGrant8, foundGrant9)
	}
}

func TestCheckPrivateForumInconsistenciesThreadPermissionsMissingParentView(t *testing.T) {
	const (
		userID   int32 = 7
		topicID1 int32 = 11
		topicID2 int32 = 12
		threadID int32 = 55
	)
	queries := &privateForumCheckQuerier{
		grants: []*db.AdminListAllPrivateForumGrantsRow{
			// User has view on topic 11 only
			{
				ID:       1,
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				Action:   consts.PermissionActionView.String(),
				ItemID:   sql.NullInt32{Int32: topicID1, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
			// User has edit grant on thread 55, but thread 55 belongs to topic 12 (which user cannot view)
			{
				ID:       2,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionEdit.String(),
				ItemID:   sql.NullInt32{Int32: threadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
			// User has edit-any grant on thread 55
			{
				ID:       3,
				Section:  consts.PermissionSectionPrivateForumThread.String(),
				Item:     sql.NullString{String: consts.PermissionItemThread.String(), Valid: true},
				Action:   consts.PermissionActionEditAny.String(),
				ItemID:   sql.NullInt32{Int32: threadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
		},
		threads: []*db.AdminListAllPrivateForumThreadsRow{
			{Idforumthread: threadID, Idforumtopic: topicID2},
		},
	}
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	inconsistencies, err := cd.CheckAndFixPrivateForumInconsistencies(context.Background(), nil, true)
	if err != nil {
		t.Fatalf("CheckAndFixPrivateForumInconsistencies: %v", err)
	}

	if len(inconsistencies) != 2 {
		t.Fatalf("expected 2 inconsistencies for thread grants missing parent view, got %d: %+v", len(inconsistencies), inconsistencies)
	}
	for _, inc := range inconsistencies {
		if inc.GrantID != 2 && inc.GrantID != 3 {
			t.Errorf("unexpected inconsistency for grant %d", inc.GrantID)
		}
	}
}

func TestCheckPrivateForumInconsistenciesOrphanedThreadAndFixExecution(t *testing.T) {
	const (
		userID           int32 = 7
		topicID          int32 = 11
		orphanedThreadID int32 = 999
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
				Action:   consts.PermissionActionEdit.String(),
				ItemID:   sql.NullInt32{Int32: orphanedThreadID, Valid: true},
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Username: sql.NullString{String: "participant", Valid: true},
			},
		},
		threads: []*db.AdminListAllPrivateForumThreadsRow{
			{Idforumthread: 10, Idforumtopic: topicID},
		},
	}
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	// Dry run first
	inconsistencies, err := cd.CheckAndFixPrivateForumInconsistencies(context.Background(), nil, true)
	if err != nil {
		t.Fatalf("CheckAndFixPrivateForumInconsistencies: %v", err)
	}
	if len(inconsistencies) != 1 {
		t.Fatalf("expected 1 orphaned thread inconsistency, got %d: %+v", len(inconsistencies), inconsistencies)
	}
	if inconsistencies[0].GrantID != 2 {
		t.Errorf("expected grant 2 flagged, got %d", inconsistencies[0].GrantID)
	}
	if len(queries.deletedGrants) != 0 {
		t.Errorf("dryRun deleted grants unexpectedly: %v", queries.deletedGrants)
	}

	// Now execute fix
	fixIDs := []string{inconsistencies[0].ID}
	fixed, err := cd.CheckAndFixPrivateForumInconsistencies(context.Background(), fixIDs, false)
	if err != nil {
		t.Fatalf("fix error: %v", err)
	}
	if len(fixed) != 1 {
		t.Errorf("expected 1 fixed inconsistency, got %d", len(fixed))
	}
	if len(queries.deletedGrants) != 1 || queries.deletedGrants[0] != 2 {
		t.Errorf("expected grant 2 deleted, got %v", queries.deletedGrants)
	}
}
