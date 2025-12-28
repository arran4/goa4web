package common

import (
	"context"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestQuerierFakeGrantStubs(t *testing.T) {
	q := &QuerierFake{}
	if _, err := q.SystemCheckGrant(context.Background(), db.SystemCheckGrantParams{}); err != nil {
		t.Fatalf("default SystemCheckGrant returned error: %v", err)
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant call, got %d", len(q.SystemCheckGrantCalls))
	}

	q.SystemCheckGrantErr = errors.New("deny")
	if _, err := q.SystemCheckGrant(context.Background(), db.SystemCheckGrantParams{Action: "view"}); err == nil {
		t.Fatalf("expected injected grant error")
	}
	if len(q.SystemCheckGrantCalls) != 2 {
		t.Fatalf("expected 2 grant calls, got %d", len(q.SystemCheckGrantCalls))
	}

	if _, err := q.SystemCheckRoleGrant(context.Background(), db.SystemCheckRoleGrantParams{}); err != nil {
		t.Fatalf("default role grant returned error: %v", err)
	}
	if len(q.SystemCheckRoleGrantCalls) != 1 {
		t.Fatalf("expected 1 role grant call, got %d", len(q.SystemCheckRoleGrantCalls))
	}

	q.SystemCheckRoleGrantErr = errors.New("deny role")
	if _, err := q.SystemCheckRoleGrant(context.Background(), db.SystemCheckRoleGrantParams{Action: "post"}); err == nil {
		t.Fatalf("expected injected role grant error")
	}
	if len(q.SystemCheckRoleGrantCalls) != 2 {
		t.Fatalf("expected 2 role grant calls, got %d", len(q.SystemCheckRoleGrantCalls))
	}
}

func TestQuerierFakeTopicListing(t *testing.T) {
	q := &QuerierFake{
		AdminListTopicsWithUserGrantsNoRolesRows: []*db.AdminListTopicsWithUserGrantsNoRolesRow{
			{Idforumtopic: 1},
		},
	}

	rows, err := q.AdminListTopicsWithUserGrantsNoRoles(context.Background(), true)
	if err != nil {
		t.Fatalf("AdminListTopicsWithUserGrantsNoRoles error: %v", err)
	}
	if len(rows) != 1 || rows[0].Idforumtopic != 1 {
		t.Fatalf("unexpected topics returned: %+v", rows)
	}
	if len(q.AdminListTopicsWithUserGrantsNoRolesCalls) != 1 {
		t.Fatalf("expected 1 topic listing call, got %d", len(q.AdminListTopicsWithUserGrantsNoRolesCalls))
	}
	if include, ok := q.AdminListTopicsWithUserGrantsNoRolesCalls[0].(bool); !ok || !include {
		t.Fatalf("expected includeAdmin=true, got %#v", q.AdminListTopicsWithUserGrantsNoRolesCalls[0])
	}
}
