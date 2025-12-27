package db

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestQueries_AdminListAllUserIDs(t *testing.T) {
	q := &QuerierStub{
		AdminListAllUserIDsReturns: []int32{1, 2},
	}
	res, err := q.AdminListAllUserIDs(context.Background())
	if err != nil {
		t.Fatalf("AdminListAllUserIDs: %v", err)
	}
	if len(res) != 2 || res[0] != 1 || res[1] != 2 {
		t.Fatalf("unexpected result %v", res)
	}
	if q.AdminListAllUserIDsCalls != 1 {
		t.Fatalf("expected 1 call, got %d", q.AdminListAllUserIDsCalls)
	}
}

func TestQueries_AdminListAllUsers(t *testing.T) {
	q := &QuerierStub{
		AdminListAllUsersReturns: []*AdminListAllUsersRow{
			{
				Idusers:  1,
				Username: sql.NullString{String: "bob", Valid: true},
			},
		},
	}
	res, err := q.AdminListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("AdminListAllUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}
	if q.AdminListAllUsersCalls != 1 {
		t.Fatalf("expected 1 call, got %d", q.AdminListAllUsersCalls)
	}
}

func TestQueries_SystemListAllUsers(t *testing.T) {
	now := time.Now()
	q := &QuerierStub{
		SystemListAllUsersReturns: []*SystemListAllUsersRow{
			{
				Idusers:   1,
				Username:  sql.NullString{String: "bob", Valid: true},
				Admin:     false,
				CreatedAt: now,
			},
		},
	}
	res, err := q.SystemListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("SystemListAllUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}
	if q.SystemListAllUsersCalls != 1 {
		t.Fatalf("expected 1 call, got %d", q.SystemListAllUsersCalls)
	}
}

func TestQueries_AdminDeleteUserByID(t *testing.T) {
	q := &QuerierStub{}

	if err := q.AdminDeleteUserByID(context.Background(), 1); err != nil {
		t.Fatalf("AdminDeleteUserByID: %v", err)
	}
	if len(q.AdminDeleteUserByIDCalls) != 1 || q.AdminDeleteUserByIDCalls[0] != 1 {
		t.Fatalf("unexpected calls %+v", q.AdminDeleteUserByIDCalls)
	}
}

func TestQueries_AdminUpdateUsernameByID(t *testing.T) {
	q := &QuerierStub{}
	params := AdminUpdateUsernameByIDParams{
		Username: sql.NullString{String: "bob", Valid: true},
		Idusers:  1,
	}
	if err := q.AdminUpdateUsernameByID(context.Background(), params); err != nil {
		t.Fatalf("AdminUpdateUsernameByID: %v", err)
	}
	if len(q.AdminUpdateUsernameByIDCalls) != 1 || q.AdminUpdateUsernameByIDCalls[0] != params {
		t.Fatalf("unexpected calls %+v", q.AdminUpdateUsernameByIDCalls)
	}
}
