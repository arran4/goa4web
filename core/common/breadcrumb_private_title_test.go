package common

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestPrivateForumBreadcrumbUsesDisplayTitle(t *testing.T) {
	queries := &db.QuerierStub{
		GetForumTopicByIdForUserReturns: map[int32]*db.GetForumTopicByIdForUserRow{
			1: {Idforumtopic: 1, Handler: "private", Title: sql.NullString{String: "Hidden", Valid: true}},
		},
		ListPrivateTopicParticipantsByTopicIDForUserReturns: map[int32][]*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
			1: {
				{Idusers: 2, Username: sql.NullString{String: "Alice", Valid: true}},
				{Idusers: 3, Username: sql.NullString{String: "Bob", Valid: true}},
			},
		},
	}
	cd := NewTestCoreData(t, queries)
	cd.SetCurrentSection("privateforum")
	cd.ForumBasePath = "/private"
	cd.UserID = 1
	cd.currentTopicID = 1

	crumbs, err := cd.forumBreadcrumbs()
	if err != nil {
		t.Fatalf("forumBreadcrumbs error: %v", err)
	}
	if len(crumbs) < 2 {
		t.Fatalf("expected >=2 crumbs, got %v", crumbs)
	}
	if crumbs[0].Title != "Private" {
		t.Fatalf("unexpected root crumb title: %v", crumbs[0].Title)
	}
	if crumbs[1].Title != "Alice, Bob" {
		t.Fatalf("unexpected crumb title: %v", crumbs[1].Title)
	}
	if len(queries.GetForumTopicByIdForUserCalls) != 1 {
		t.Fatalf("expected single topic lookup, got %d", len(queries.GetForumTopicByIdForUserCalls))
	}
	if len(queries.ListPrivateTopicParticipantsByTopicIDForUserCalls) != 1 {
		t.Fatalf("expected single participant lookup, got %d", len(queries.ListPrivateTopicParticipantsByTopicIDForUserCalls))
	}
}
