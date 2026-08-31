package scenario

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestApplyUserCreateAndEnable(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: user-slice-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: secret-password
At: 2026-08-01T09:00:00Z

-- 02-enable.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:02:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 1. SystemInsertUser expectation
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnResult(sqlmock.NewResult(15, 1))

	// 2. InsertUserEmail expectation (standard unverified email from CreateUserWithEmail)
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(int32(15), "alice@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. InsertPassword expectation
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(int32(15), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. SystemCreateUserRole expectation (standard ApproveUser)
	mock.ExpectExec("(?s).*SystemCreateUserRole.*").
		WithArgs(int32(15), "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := runner.Apply(ctx, sc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if res.EventsApplied != 2 {
		t.Errorf("expected 2 events applied, got %d", res.EventsApplied)
	}

	uid, ok := res.Registry.ResolveUser("alice")
	if !ok || uid != 15 {
		t.Errorf("expected alice to resolve to 15, got %d (ok=%v)", uid, ok)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestApplyPreflightRejectsUnsupportedOperationWithoutMutation(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	// Scenario containing valid format operations, but private-forum.create is not supported by the runner
	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: partial-apply-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
At: 2026-08-01T09:00:00Z

-- 02-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Title: Staff Room
At: 2026-08-01T09:05:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Crucial assertion: No mock expectations are set.
	// If any DB query runs, sqlmock will fail with unexpected query error.
	res, err := runner.Apply(ctx, sc)
	if err == nil {
		t.Fatal("expected Apply to fail during preflight, but it succeeded")
	}

	var errUnsupported ErrUnsupportedOperation
	if !errors.As(err, &errUnsupported) {
		t.Fatalf("expected ErrUnsupportedOperation, got: %v", err)
	}
	if errUnsupported.Op != "private-forum.create" {
		t.Errorf("expected unsupported op 'private-forum.create', got %q", errUnsupported.Op)
	}
	if res != nil {
		t.Errorf("expected nil ApplyResult on failure, got: %+v", res)
	}

	// Verify ZERO queries were executed
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls made: %v", err)
	}
}
