//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/testdata/scenarios"
)

func TestScenarioServeCmd_ParseCLI(t *testing.T) {
	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"serve"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	// 1. Missing path error
	serveCmdEmpty, err := parseScenarioServeCmd(parent, []string{})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if err := serveCmdEmpty.Run(); err == nil || !strings.Contains(err.Error(), "scenario path required") {
		t.Fatalf("expected scenario path required error, got: %v", err)
	}

	// 2. Directory path argument
	serveCmdDir, err := parseScenarioServeCmd(parent, []string{"testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmdDir.Path != "testdata/scenarios/100-private-forum" {
		t.Errorf("expected path testdata/scenarios/100-private-forum, got %q", serveCmdDir.Path)
	}

	// 3. Direct file path argument
	serveCmdFile, err := parseScenarioServeCmd(parent, []string{"testdata/scenarios/100-private-forum/scenario.txtar"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmdFile.Path != "testdata/scenarios/100-private-forum/scenario.txtar" {
		t.Errorf("expected path testdata/scenarios/100-private-forum/scenario.txtar, got %q", serveCmdFile.Path)
	}

	// 4. Custom listen flag on subcommand
	serveCmdListen, err := parseScenarioServeCmd(parent, []string{"--listen", ":9090", "testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmdListen.Listen != ":9090" {
		t.Errorf("expected listen :9090, got %q", serveCmdListen.Listen)
	}
	if serveCmdListen.Path != "testdata/scenarios/100-private-forum" {
		t.Errorf("expected path testdata/scenarios/100-private-forum, got %q", serveCmdListen.Path)
	}
}

func TestScenarioServeCmd_BootstrapAndSameDatabaseInvariant(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "--listen", ":8082", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	// 1. Verify same-database invariant: the database connection in srv is dbConn
	if srv.DB != dbConn {
		t.Errorf("server DB (%p) does not match bootstrap DB (%p)", srv.DB, dbConn)
	}

	// 2. Verify schema version is migrated to 97
	var currentVersion int64
	err = dbConn.QueryRowContext(ctx, "SELECT version_id FROM goose_db_version ORDER BY id DESC LIMIT 1;").Scan(&currentVersion)
	if err != nil {
		t.Fatalf("query goose_db_version: %v", err)
	}
	if currentVersion != 97 {
		t.Errorf("expected migrated schema version 97, got %d", currentVersion)
	}

	// 3. Verify language English is seeded
	var langName string
	err = dbConn.QueryRowContext(ctx, "SELECT nameof FROM language WHERE id = 1;").Scan(&langName)
	if err != nil {
		t.Fatalf("query language: %v", err)
	}
	if langName != "English" {
		t.Errorf("expected language English, got %q", langName)
	}

	// 4. Verify users Alice and Bob exist and have correct passwords
	var aliceUID, bobUID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'alice';").Scan(&aliceUID)
	if err != nil {
		t.Fatalf("query alice idusers: %v", err)
	}
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'bob';").Scan(&bobUID)
	if err != nil {
		t.Fatalf("query bob idusers: %v", err)
	}

	var alicePasswd, aliceAlg string
	err = dbConn.QueryRowContext(ctx, "SELECT passwd, passwd_algorithm FROM passwords WHERE users_idusers = ?;", aliceUID).Scan(&alicePasswd, &aliceAlg)
	if err != nil {
		t.Fatalf("query alice password: %v", err)
	}
	if !auth.VerifyPassword("alice-test", alicePasswd, aliceAlg) {
		t.Errorf("alice password verification failed for 'alice-test'")
	}

	var bobPasswd, bobAlg string
	err = dbConn.QueryRowContext(ctx, "SELECT passwd, passwd_algorithm FROM passwords WHERE users_idusers = ?;", bobUID).Scan(&bobPasswd, &bobAlg)
	if err != nil {
		t.Fatalf("query bob password: %v", err)
	}
	if !auth.VerifyPassword("bob-test", bobPasswd, bobAlg) {
		t.Errorf("bob password verification failed for 'bob-test'")
	}

	// 5. Verify private topic "Staff Room" was created
	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, srv.Config)
	aliceCD := cd.ForUser(aliceUID)
	bobCD := cd.ForUser(bobUID)

	// Alice has global see and create; Bob has global see (not create)
	if !aliceCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Error("expected Alice to have global privateforum create permission")
	}
	if !aliceCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Error("expected Alice to have global privateforum see permission")
	}
	if !bobCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Error("expected Bob to have global privateforum see permission")
	}
	if bobCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Error("expected Bob NOT to have global privateforum create permission")
	}

	// Verify topic
	var topicID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room';").Scan(&topicID)
	if err != nil {
		t.Fatalf("query Staff Room topic: %v", err)
	}
	topic, err := querier.GetForumTopicById(ctx, topicID)
	if err != nil {
		t.Fatalf("GetForumTopicById: %v", err)
	}
	if topic.Title.String != "Staff Room" {
		t.Errorf("expected topic title 'Staff Room', got %q", topic.Title.String)
	}
	if topic.Description.String != "Private discussion for Alice and Bob" {
		t.Errorf("expected topic description 'Private discussion for Alice and Bob', got %q", topic.Description.String)
	}

	// Verify topic-specific participant grants
	for _, act := range []string{"see", "view", "post", "reply"} {
		if !aliceCD.HasGrant("privateforum", "topic", act, topicID) {
			t.Errorf("expected Alice to have grant %s on private topic %d", act, topicID)
		}
		if !bobCD.HasGrant("privateforum", "topic", act, topicID) {
			t.Errorf("expected Bob to have grant %s on private topic %d", act, topicID)
		}
	}

	// 6. Security boundary: verify an unrelated user has NO global privateforum grants
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
		t.Fatal("security violation: baseline user unexpectedly has global privateforum create permission")
	}
	if baselineCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Fatal("security violation: baseline user unexpectedly has global privateforum see permission")
	}
}

func TestScenarioServeCmd_IsolationFromConfiguredDBConn(t *testing.T) {
	ctx := context.Background()

	// Configure root with an intentionally broken / unreachable DB_CONN and MySQL driver
	root, err := parseRoot([]string{"goa4web", "--db-conn", "mysql://invalid-nonexistent-host:9999/dummy", "--db-driver", "mysql", "scenario", "serve", "scenarios/valid"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	fsys := fstest.MapFS{
		"scenarios/valid/scenario.txtar": &fstest.MapFile{
			Data: []byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: isolation-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z
`),
		},
	}

	parent, err := parseScenarioCmd(root, []string{"serve", "scenarios/valid"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"scenarios/valid"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = fsys

	// Bootstrap should succeed because it creates its own ephemeral SQLite DB, completely ignoring the broken DB_CONN
	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap should succeed in isolation from DB_CONN, but failed: %v", err)
	}
	defer cleanup()

	// Verify the database used is SQLite, not MySQL
	if srv.Config.DBDriver != "sqlite3" {
		t.Errorf("expected DBDriver sqlite3, got %q", srv.Config.DBDriver)
	}
	if srv.Config.DBConn != "" {
		t.Errorf("expected DBConn to be cleared for ephemeral serve, got %q", srv.Config.DBConn)
	}

	// Verify Alice is created in the ephemeral SQLite database
	var aliceCount int
	if err := dbConn.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE username = 'alice';").Scan(&aliceCount); err != nil {
		t.Fatalf("query alice in ephemeral sqlite DB: %v", err)
	}
	if aliceCount != 1 {
		t.Errorf("expected 1 alice user in ephemeral DB, got %d", aliceCount)
	}
}

func TestScenarioServeCmd_HTTPSmokeTest(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, _, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	// Test 1: HTTP GET / returns 200 OK
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET / returned status %d, want %d", rec.Code, http.StatusOK)
	}

	// Test 2: HTTP GET /login returns 200 OK
	loginReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRec := httptest.NewRecorder()
	srv.Router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Errorf("GET /login returned status %d, want %d", loginRec.Code, http.StatusOK)
	}

	// Test 3: HTTP GET /privateforum/ (requires login, redirects or returns 200/403/302 appropriately)
	pfReq := httptest.NewRequest(http.MethodGet, "/privateforum/", nil)
	pfRec := httptest.NewRecorder()
	srv.Router.ServeHTTP(pfRec, pfReq)

	// An unauthenticated user accessing private forum gets redirected to login or denied
	if pfRec.Code != http.StatusOK && pfRec.Code != http.StatusFound && pfRec.Code != http.StatusForbidden && pfRec.Code != http.StatusSeeOther {
		t.Errorf("GET /privateforum/ returned unexpected status %d", pfRec.Code)
	}
}
