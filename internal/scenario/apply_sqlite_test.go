//go:build sqlite || sqlite3

package scenario

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sqlutil"
	"github.com/arran4/goa4web/migrations"
)

func TestApplyScenarioWithRealSQLiteDB(t *testing.T) {
	ctx := context.Background()

	dbConn, err := sql.Open("sqlite", "file:scenario_real_sqlite_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("sql.Open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = dbConn.Close()
	})

	// 1. Apply schema migrations in-memory
	if err := dbstart.Apply(ctx, dbConn, migrations.FS, false, "sqlite3"); err != nil {
		t.Fatalf("dbstart.Apply: %v", err)
	}

	// 2. Apply database seed SQL (roles, role grants, etc.)
	seedSQL := database.SeedSQLForDriver("sqlite3")
	if err := sqlutil.RunStatements(ctx, dbConn, strings.NewReader(string(seedSQL))); err != nil {
		t.Fatalf("run seed SQL: %v", err)
	}

	// 3. Seed default language
	if _, err := dbConn.ExecContext(ctx, `INSERT INTO language (id, nameof) VALUES (1, 'English');`); err != nil {
		t.Fatalf("insert default language: %v", err)
	}

	querier := db.NewForDriver(dbConn, "sqlite3")
	cfg := &config.RuntimeConfig{DBDriver: "sqlite3"}
	cd := common.NewCoreData(ctx, querier, cfg)

	// Step A: Security boundary verification before scenario apply.
	// Create a baseline user with 'user' role and verify that on a freshly seeded database without the scenario,
	// normal users with the 'user' role DO NOT implicitly possess privateforum see/create grants.
	initUserRes, err := querier.SystemInsertUser(ctx, sql.NullString{String: "baseline_user", Valid: true})
	if err != nil {
		t.Fatalf("insert baseline user: %v", err)
	}
	baselineUID := int32(initUserRes)
	if err := querier.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{
		UsersIdusers: baselineUID,
		Name:         "user",
	}); err != nil {
		t.Fatalf("assign user role: %v", err)
	}

	baselineCD := cd.ForUser(baselineUID)
	if baselineCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Fatal("security violation: baseline user unexpectedly has global privateforum create permission before scenario apply")
	}
	if baselineCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Fatal("security violation: baseline user unexpectedly has global privateforum see permission before scenario apply")
	}

	// Step B: Self-provisioning scenario.
	// Scenario explicitly defines users, grants required privateforum permissions directly to Alice and Bob via user.grant,
	// and creates the private forum between Alice and Bob.
	runner := NewRunner(cd)

	scenarioTxt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: real-db-private-forum
Description: Real in-memory SQLite integration test for permission provisioning and private topic creation

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-password123
At: 2026-08-01T09:00:00Z

-- 02-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:01:00Z

-- 022-grant-alice-see.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:01:30Z

-- 024-grant-alice-create.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: create
At: 2026-08-01T09:01:45Z

-- 03-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-password456
At: 2026-08-01T09:02:00Z

-- 04-enable-bob.event --
Op: user.enable
Actor: admin
User: bob
At: 2026-08-01T09:03:00Z

-- 042-grant-bob-see.event --
Op: user.grant
User: bob
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:03:30Z

-- 05-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
Description: Private discussion between Alice and Bob
At: 2026-08-01T09:05:00Z
`

	sc, err := Parse([]byte(scenarioTxt), nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	res, err := runner.Apply(ctx, sc)
	if err != nil {
		t.Fatalf("runner.Apply failed: %v", err)
	}

	if res.EventsApplied != 8 {
		t.Errorf("expected 8 events applied, got %d", res.EventsApplied)
	}

	// Step C: Verify that baseline user STILL has no privateforum permissions after scenario apply
	if baselineCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Fatal("security violation: baseline user unexpectedly gained privateforum create permission after scenario apply")
	}
	if baselineCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Fatal("security violation: baseline user unexpectedly gained privateforum see permission after scenario apply")
	}

	aliceUID, ok := res.Registry.ResolveUser("alice")
	if !ok || aliceUID == 0 {
		t.Fatalf("failed to resolve alice user ref: %d (ok=%v)", aliceUID, ok)
	}

	bobUID, ok := res.Registry.ResolveUser("bob")
	if !ok || bobUID == 0 {
		t.Fatalf("failed to resolve bob user ref: %d (ok=%v)", bobUID, ok)
	}

	aliceCD := cd.ForUser(aliceUID)
	bobCD := cd.ForUser(bobUID)

	// Verify that the permissions are now established specifically for Alice and Bob
	if !aliceCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Error("expected Alice to have global privateforum create permission after user.grant")
	}
	if !aliceCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Error("expected Alice to have global privateforum see permission after user.grant")
	}
	if !bobCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Error("expected Bob to have global privateforum see permission after user.grant")
	}
	if bobCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Error("expected Bob NOT to have global privateforum create permission since it was not granted")
	}

	topicIDVal, ok := res.Registry.Resolve(RefTypeForum, "staff-room")
	if !ok || topicIDVal == nil {
		t.Fatalf("failed to resolve staff-room forum ref (ok=%v)", ok)
	}
	topicID, ok := topicIDVal.(int32)
	if !ok || topicID == 0 {
		t.Fatalf("expected int32 topic ID, got %T (%v)", topicIDVal, topicIDVal)
	}

	// Verify topic exists in real database
	topic, err := querier.GetForumTopicById(ctx, topicID)
	if err != nil {
		t.Fatalf("GetForumTopicById: %v", err)
	}
	if topic.Title.String != "Staff Room" {
		t.Errorf("topic title = %q, want 'Staff Room'", topic.Title.String)
	}
	if topic.Description.String != "Private discussion between Alice and Bob" {
		t.Errorf("topic description = %q, want 'Private discussion between Alice and Bob'", topic.Description.String)
	}

	// Verify Alice permissions on the private topic
	for _, act := range []string{"see", "view", "post", "reply"} {
		if !aliceCD.HasGrant("privateforum", "topic", act, topicID) {
			t.Errorf("expected Alice to have grant %s on private topic %d", act, topicID)
		}
	}

	// Verify Bob permissions on the private topic
	for _, act := range []string{"see", "view", "post", "reply"} {
		if !bobCD.HasGrant("privateforum", "topic", act, topicID) {
			t.Errorf("expected Bob to have grant %s on private topic %d", act, topicID)
		}
	}

	// Verify passwords in real database
	var alicePasswd, aliceAlg string
	err = dbConn.QueryRowContext(ctx, `SELECT passwd, passwd_algorithm FROM passwords WHERE users_idusers = ?;`, aliceUID).Scan(&alicePasswd, &aliceAlg)
	if err != nil {
		t.Fatalf("query alice password: %v", err)
	}
	if !auth.VerifyPassword("alice-password123", alicePasswd, aliceAlg) {
		t.Error("alice password hash verification failed for 'alice-password123'")
	}

	var bobPasswd, bobAlg string
	err = dbConn.QueryRowContext(ctx, `SELECT passwd, passwd_algorithm FROM passwords WHERE users_idusers = ?;`, bobUID).Scan(&bobPasswd, &bobAlg)
	if err != nil {
		t.Fatalf("query bob password: %v", err)
	}
	if !auth.VerifyPassword("bob-password456", bobPasswd, bobAlg) {
		t.Error("bob password hash verification failed for 'bob-password456'")
	}
}

func TestSeedSQLDoesNotContainPrivateForumGrants(t *testing.T) {
	for _, driver := range []string{"mysql", "sqlite3"} {
		seed := string(database.SeedSQLForDriver(driver))
		if strings.Contains(seed, "privateforum") {
			t.Errorf("seed SQL for driver %q must not contain privateforum grants", driver)
		}
	}
}
