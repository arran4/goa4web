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

func TestApplyUserCreateAndEnableAndPrivateForumCreate(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := conn.Close(); err != nil {
			t.Errorf("conn.Close: %v", err)
		}
	})

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: private-forum-full-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-test
At: 2026-08-01T09:00:00Z

-- 02-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:02:00Z

-- 022-grant-alice-see.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:02:30Z

-- 024-grant-alice-create.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: create
At: 2026-08-01T09:02:45Z

-- 03-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-test
At: 2026-08-01T09:03:00Z

-- 04-enable-bob.event --
Op: user.enable
Actor: admin
User: bob
At: 2026-08-01T09:04:00Z

-- 042-grant-bob-see.event --
Op: user.grant
User: bob
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:04:30Z

-- 05-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
Description: Staff discussion
At: 2026-08-01T09:10:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	aliceUID := int32(10)
	bobUID := int32(20)
	topicID := int32(100)

	// 1. Alice creation
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnResult(sqlmock.NewResult(int64(aliceUID), 1))
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(aliceUID, "alice@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(aliceUID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Alice enable
	mock.ExpectExec("(?s).*SystemCreateUserRole.*").
		WithArgs(aliceUID, "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Alice user grants
	mock.ExpectExec("(?s).*AdminCreateGrant.*").
		WithArgs(
			sql.NullInt32{Int32: aliceUID, Valid: true},
			sql.NullInt32{},
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"allow",
			sql.NullInt32{},
			sql.NullString{},
			"see",
			sql.NullString{},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("(?s).*AdminCreateGrant.*").
		WithArgs(
			sql.NullInt32{Int32: aliceUID, Valid: true},
			sql.NullInt32{},
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"allow",
			sql.NullInt32{},
			sql.NullString{},
			"create",
			sql.NullString{},
		).
		WillReturnResult(sqlmock.NewResult(2, 1))

	// 4. Bob creation
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "bob", Valid: true}).
		WillReturnResult(sqlmock.NewResult(int64(bobUID), 1))
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(bobUID, "bob@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(bobUID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Bob enable
	mock.ExpectExec("(?s).*SystemCreateUserRole.*").
		WithArgs(bobUID, "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 6. Bob user grant
	mock.ExpectExec("(?s).*AdminCreateGrant.*").
		WithArgs(
			sql.NullInt32{Int32: bobUID, Valid: true},
			sql.NullInt32{},
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"allow",
			sql.NullInt32{},
			sql.NullString{},
			"see",
			sql.NullString{},
		).
		WillReturnResult(sqlmock.NewResult(3, 1))

	// 7. Private forum creation by Alice with Bob
	// Check creator grant
	mock.ExpectQuery("(?s).*SystemCheckGrant.*").
		WithArgs(aliceUID, "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, false, sql.NullInt32{Int32: aliceUID, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// Check participant eligibility for Bob
	mock.ExpectQuery("(?s).*SystemCheckGrant.*").
		WithArgs(bobUID, "privateforum", sql.NullString{String: "topic", Valid: true}, "see", sql.NullInt32{}, false, sql.NullInt32{Int32: bobUID, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// CreateForumTopicForPoster
	mock.ExpectExec("(?s).*CreateForumTopicForPoster.*").
		WithArgs(
			int32(0), // PrivateForumCategoryID
			sql.NullInt32{Int32: 1, Valid: true},
			sql.NullString{String: "Staff Room", Valid: true},
			sql.NullString{String: "Staff discussion", Valid: true},
			"private",
			"privateforum",
			sql.NullInt32{Int32: 0, Valid: true},
			sql.NullInt32{Int32: aliceUID, Valid: true},
			aliceUID,
		).WillReturnResult(sqlmock.NewResult(int64(topicID), 1))

	// Grants and subscriptions for Alice & Bob
	for _, uid := range []int32{bobUID, aliceUID} {
		for _, act := range []string{"see", "view", "post", "reply"} {
			mock.ExpectExec("(?s).*INSERT INTO grants.*").
				WithArgs(
					sql.NullInt32{Int32: uid, Valid: true},
					sql.NullInt32{},
					"privateforum",
					sql.NullString{String: "topic", Valid: true},
					"allow",
					sql.NullInt32{Int32: topicID, Valid: true},
					sql.NullString{},
					act,
					sql.NullString{},
				).WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectExec("(?s).*InsertSubscription.*").
			WithArgs(uid, "create thread:/forum/topic/100/*", "internal").
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	res, err := runner.Apply(ctx, sc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if res.EventsApplied != 8 {
		t.Errorf("expected 8 events applied, got %d", res.EventsApplied)
	}

	resolvedAlice, ok := res.Registry.ResolveUser("alice")
	if !ok || resolvedAlice != aliceUID {
		t.Errorf("expected alice to resolve to %d, got %d (ok=%v)", aliceUID, resolvedAlice, ok)
	}

	resolvedBob, ok := res.Registry.ResolveUser("bob")
	if !ok || resolvedBob != bobUID {
		t.Errorf("expected bob to resolve to %d, got %d (ok=%v)", bobUID, resolvedBob, ok)
	}

	resolvedTopic, ok := res.Registry.Resolve(RefTypeForum, "staff-room")
	if !ok || resolvedTopic != topicID {
		t.Errorf("expected staff-room to resolve to %d, got %d (ok=%v)", topicID, resolvedTopic, ok)
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
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := conn.Close(); err != nil {
			t.Errorf("conn.Close: %v", err)
		}
	})

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	// Scenario containing valid format operations, but forum.post is not supported by the runner
	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: partial-apply-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-password
At: 2026-08-01T09:00:00Z

-- 02-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-password
At: 2026-08-01T09:01:00Z

-- 03-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
At: 2026-08-01T09:02:00Z

-- 04-post.event --
Op: forum.post
Ref: welcome-post
Actor: alice
Forum: staff-room
At: 2026-08-01T09:05:00Z

Post content here.
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
	if errUnsupported.Op != "forum.post" {
		t.Errorf("expected unsupported op 'forum.post', got %q", errUnsupported.Op)
	}
	if res != nil {
		t.Errorf("expected nil ApplyResult on failure, got: %+v", res)
	}

	// Verify ZERO queries were executed
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls made: %v", err)
	}
}

func TestApplyPreflightRejectsUnboundAdminActorInPrivateForumWithZeroMutation(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := conn.Close(); err != nil {
			t.Errorf("conn.Close: %v", err)
		}
	})

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	// Scenario where Actor: admin is used in private-forum.create without admin being declared as a user
	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: unbound-admin-actor-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-password
At: 2026-08-01T09:00:00Z

-- 02-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: admin
Participant: alice
Title: Staff Room
At: 2026-08-01T09:05:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Zero mock queries set - preflight must reject it before touching the DB
	res, err := runner.Apply(ctx, sc)
	if err == nil {
		t.Fatal("expected Apply to fail during preflight for unbound admin actor, got success")
	}

	var errUnresolved ErrUnresolvedRef
	if !errors.As(err, &errUnresolved) {
		t.Fatalf("expected ErrUnresolvedRef, got: %v", err)
	}
	if errUnresolved.Symbol != "admin" || errUnresolved.Field != "Actor" {
		t.Errorf("unexpected unresolved ref details: %+v", errUnresolved)
	}
	if res != nil {
		t.Errorf("expected nil ApplyResult, got: %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls made: %v", err)
	}
}

func TestApplyRejectsIneligibleParticipant(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := conn.Close(); err != nil {
			t.Errorf("conn.Close: %v", err)
		}
	})

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: ineligible-participant-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-pass
At: 2026-08-01T09:00:00Z

-- 02-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-pass
At: 2026-08-01T09:01:00Z

-- 03-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
At: 2026-08-01T09:05:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	aliceUID := int32(10)
	bobUID := int32(20)

	// Alice creation
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnResult(sqlmock.NewResult(int64(aliceUID), 1))
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(aliceUID, "alice@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(aliceUID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Bob creation
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "bob", Valid: true}).
		WillReturnResult(sqlmock.NewResult(int64(bobUID), 1))
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(bobUID, "bob@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(bobUID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Private forum creation: Alice check passes
	mock.ExpectQuery("(?s).*SystemCheckGrant.*").
		WithArgs(aliceUID, "privateforum", sql.NullString{String: "topic", Valid: true}, "create", sql.NullInt32{}, false, sql.NullInt32{Int32: aliceUID, Valid: true}, false).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// Bob eligibility check fails
	mock.ExpectQuery("(?s).*SystemCheckGrant.*").
		WithArgs(bobUID, "privateforum", sql.NullString{String: "topic", Valid: true}, "see", sql.NullInt32{}, false, sql.NullInt32{Int32: bobUID, Valid: true}, false).
		WillReturnError(sql.ErrNoRows)

	res, err := runner.Apply(ctx, sc)
	if err == nil {
		t.Fatal("expected Apply to fail for ineligible participant, got success")
	}

	var errInvalid *common.ErrInvalidParticipants
	if !errors.As(err, &errInvalid) {
		t.Fatalf("expected ErrInvalidParticipants, got: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil ApplyResult on failure, got: %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestApplyUserGrant(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := conn.Close(); err != nil {
			t.Errorf("conn.Close: %v", err)
		}
	})

	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, &config.RuntimeConfig{})
	runner := NewRunner(cd)

	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: user-grant-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02-grant.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:01:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	aliceUID := int32(10)
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnResult(sqlmock.NewResult(int64(aliceUID), 1))
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(aliceUID, "alice@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(aliceUID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("(?s).*AdminCreateGrant.*").
		WithArgs(
			sql.NullInt32{Int32: aliceUID, Valid: true},
			sql.NullInt32{},
			"privateforum",
			sql.NullString{String: "topic", Valid: true},
			"allow",
			sql.NullInt32{},
			sql.NullString{},
			"see",
			sql.NullString{},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := runner.Apply(ctx, sc)
	if err != nil {
		t.Fatalf("runner.Apply failed: %v", err)
	}
	if res.EventsApplied != 2 {
		t.Errorf("expected 2 events applied, got %d", res.EventsApplied)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
