package scenario

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestApplyUserCreateAndEnable(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	runner := NewRunner(queries)

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

	// 2. InsertUserEmail expectation
	atTime := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(int32(15), "alice@example.test", sql.NullTime{Time: atTime, Valid: true}, sql.NullString{}, nil, 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. InsertPassword expectation
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(int32(15), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. SystemCreateUserRole expectation
	mock.ExpectExec("(?s).*SystemCreateUserRole.*").
		WithArgs(int32(15), "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
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
