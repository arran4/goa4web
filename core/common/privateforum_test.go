package common

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

type privateForumQuerierStub struct {
	db.QuerierStub
	calls []string
}

func (q *privateForumQuerierStub) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	q.calls = append(q.calls, "SystemCheckGrant")
	return q.QuerierStub.SystemCheckGrant(ctx, arg)
}

func (*privateForumQuerierStub) GetPermissionsByUserID(ctx context.Context, userID int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return nil, nil
}

func (q *privateForumQuerierStub) ListPrivateTopicsByUserID(ctx context.Context, userID sql.NullInt32) ([]*db.ListPrivateTopicsByUserIDRow, error) {
	q.calls = append(q.calls, "ListPrivateTopicsByUserID")
	return q.QuerierStub.ListPrivateTopicsByUserID(ctx, userID)
}

func (q *privateForumQuerierStub) ListPrivateTopicParticipantsByTopicIDForUser(ctx context.Context, arg db.ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*db.ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
	q.calls = append(q.calls, "ListPrivateTopicParticipantsByTopicIDForUser")
	return q.QuerierStub.ListPrivateTopicParticipantsByTopicIDForUser(ctx, arg)
}

func (q *privateForumQuerierStub) ListContentPublicLabels(ctx context.Context, arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
	q.calls = append(q.calls, "ListContentPublicLabels")
	return q.QuerierStub.ListContentPublicLabels(ctx, arg)
}

func (q *privateForumQuerierStub) ListContentLabelStatus(ctx context.Context, arg db.ListContentLabelStatusParams) ([]*db.ListContentLabelStatusRow, error) {
	q.calls = append(q.calls, "ListContentLabelStatus")
	return q.QuerierStub.ListContentLabelStatus(ctx, arg)
}

func TestCoreData_PrivateForumTopics(t *testing.T) {
	q := &privateForumQuerierStub{
		QuerierStub: db.QuerierStub{
			SystemCheckGrantReturns: 1,
		},
	}
	cd := NewTestCoreData(t, q)
	cd.UserID = 1

	viewer := sql.NullInt32{Int32: 1, Valid: true}
	topicRow := &db.ListPrivateTopicsByUserIDRow{
		Idforumtopic:                 1,
		Lastposter:                   1,
		ForumcategoryIdforumcategory: 0,
		LanguageID:                   sql.NullInt32{Int32: 1, Valid: true},
		Title:                        sql.NullString{String: "Test Topic", Valid: true},
		Description:                  sql.NullString{String: "Test Description", Valid: true},
		Threads:                      sql.NullInt32{Int32: 1, Valid: true},
		Comments:                     sql.NullInt32{Int32: 1, Valid: true},
		Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
		Handler:                      "private",
		Lastposterusername:           sql.NullString{String: "testuser", Valid: true},
	}
	q.ListPrivateTopicsByUserIDReturns = []*db.ListPrivateTopicsByUserIDRow{topicRow}

	participantParams := db.ListPrivateTopicParticipantsByTopicIDForUserParams{
		TopicID:  sql.NullInt32{Int32: 1, Valid: true},
		ViewerID: viewer,
	}
	q.ListPrivateTopicParticipantsByTopicIDForUserReturn = map[db.ListPrivateTopicParticipantsByTopicIDForUserParams][]*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
		participantParams: {{
			Idusers:  2,
			Username: sql.NullString{String: "participant1", Valid: true},
		}},
	}
	labelParams := db.ListContentPublicLabelsParams{Item: "thread", ItemID: 1}
	q.ListContentPublicLabelsReturn = map[db.ListContentPublicLabelsParams][]*db.ListContentPublicLabelsRow{
		labelParams: {{
			Item:   "thread",
			ItemID: 1,
			Label:  "public",
		}},
	}
	labelStatusParams := db.ListContentLabelStatusParams{Item: "thread", ItemID: 1}
	q.ListContentLabelStatusReturn = map[db.ListContentLabelStatusParams][]*db.ListContentLabelStatusRow{
		labelStatusParams: {{
			Item:   "thread",
			ItemID: 1,
			Label:  "owner",
		}},
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

	expectedCalls := []string{
		"SystemCheckGrant",
		"ListPrivateTopicsByUserID",
		"ListPrivateTopicParticipantsByTopicIDForUser",
		"ListContentPublicLabels",
		"ListContentLabelStatus",
	}
	if !reflect.DeepEqual(expectedCalls, q.calls) {
		t.Fatalf("unexpected call order: got %v want %v", q.calls, expectedCalls)
	}
	if len(q.SystemCheckGrantCalls) != 1 || q.SystemCheckGrantCalls[0].Section != "privateforum" {
		t.Fatalf("unexpected grant checks: %+v", q.SystemCheckGrantCalls)
	}
	if got := q.ListPrivateTopicsByUserIDCalls; len(got) != 1 || got[0] != viewer {
		t.Fatalf("unexpected topic lookups: %+v", got)
	}
	if got := q.ListPrivateTopicParticipantsByTopicIDForUserCalls; len(got) != 1 || !reflect.DeepEqual(got[0], participantParams) {
		t.Fatalf("unexpected participant lookups: %+v", got)
	}
	if got := q.ListContentPublicLabelsCalls; len(got) != 1 || !reflect.DeepEqual(got[0], labelParams) {
		t.Fatalf("unexpected public label lookups: %+v", got)
	}
	if got := q.ListContentLabelStatusCalls; len(got) != 1 || !reflect.DeepEqual(got[0], labelStatusParams) {
		t.Fatalf("unexpected label status lookups: %+v", got)
	}
}
