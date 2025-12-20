package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// BaseQuerier provides common permission helpers for test stubs.
type BaseQuerier struct {
	*db.Queries
	t                testing.TB
	GrantAllowed     bool
	RoleGrantAllowed bool
}

// NewBaseQuerier returns a base querier with default denied grants.
func NewBaseQuerier(t testing.TB) *BaseQuerier {
	t.Helper()
	return &BaseQuerier{
		Queries: db.New(panicDB{t: t}),
		t:       t,
	}
}

// AllowGrants enables SystemCheckGrant responses.
func (q *BaseQuerier) AllowGrants() {
	q.GrantAllowed = true
}

// AllowRoleGrants enables SystemCheckRoleGrant responses.
func (q *BaseQuerier) AllowRoleGrants() {
	q.RoleGrantAllowed = true
}

// SystemCheckGrant implements db.Querier.
func (q *BaseQuerier) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	if q.GrantAllowed {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}

// SystemCheckRoleGrant implements db.Querier.
func (q *BaseQuerier) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	if q.RoleGrantAllowed {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}

type panicDB struct {
	t testing.TB
}

func (p panicDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if p.t != nil {
		p.t.Helper()
		p.t.Fatalf("unexpected ExecContext: %s", query)
	}
	return nil, sql.ErrConnDone
}

func (p panicDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if p.t != nil {
		p.t.Helper()
		p.t.Fatalf("unexpected QueryContext: %s", query)
	}
	return nil, sql.ErrConnDone
}

func (p panicDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if p.t != nil {
		p.t.Helper()
		p.t.Fatalf("unexpected QueryRowContext: %s", query)
	}
	return &sql.Row{}
}

func (p panicDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if p.t != nil {
		p.t.Helper()
		p.t.Fatalf("unexpected PrepareContext: %s", query)
	}
	return nil, sql.ErrConnDone
}
