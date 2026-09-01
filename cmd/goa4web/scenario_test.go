package main

import (
	"database/sql"
	"flag"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/testdata/scenarios"
)

func TestScenarioCmd_ParseAndRun(t *testing.T) {
	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}

	// Test missing subcommand
	cmd, err := parseScenarioCmd(root, []string{})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}
	if err := cmd.Run(); err == nil || !strings.Contains(err.Error(), "missing scenario command") {
		t.Fatalf("expected missing scenario command error, got: %v", err)
	}

	// Test unknown subcommand
	cmd, err = parseScenarioCmd(root, []string{"unknown"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}
	if err := cmd.Run(); err == nil || !strings.Contains(err.Error(), "unknown scenario command") {
		t.Fatalf("expected unknown scenario command error, got: %v", err)
	}
}

func TestScenarioValidateCmd_InjectedFS(t *testing.T) {
	fsys := fstest.MapFS{
		"scenarios/valid/scenario.txtar": &fstest.MapFile{
			Data: []byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: cli-test

-- 01-user.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-pass
At: 2026-08-01T09:00:00Z

-- 02-enable.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:01:00Z
`),
		},
		"scenarios/valid/assets/img.jpg": &fstest.MapFile{
			Data: []byte("asset-data"),
		},
		"scenarios/missing-asset/scenario.txtar": &fstest.MapFile{
			Data: []byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: missing-asset-test

-- 01-user.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-pass
At: 2026-08-01T09:00:00Z

-- 02-user-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-pass
At: 2026-08-01T09:01:00Z

-- 03-forum.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Participant: bob
Title: Forum 1
At: 2026-08-01T09:02:00Z

-- 04-post.event --
Op: forum.post
Ref: post1
Actor: alice
Forum: forum1
Attachment: assets/notfound.jpg
At: 2026-08-01T09:05:00Z
`),
		},
		"scenarios/no-txtar/dummy.txt": &fstest.MapFile{
			Data: []byte("dummy"),
		},
	}

	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"validate"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	// 1. Validate by file path
	valCmdFile, err := parseScenarioValidateCmd(parent, []string{"scenarios/valid/scenario.txtar"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmdFile.fsys = fsys
	if err := valCmdFile.Run(); err != nil {
		t.Fatalf("expected validation success for file, got: %v", err)
	}

	// 2. Validate by directory path
	valCmdDir, err := parseScenarioValidateCmd(parent, []string{"scenarios/valid"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmdDir.fsys = fsys
	if err := valCmdDir.Run(); err != nil {
		t.Fatalf("expected validation success for directory, got: %v", err)
	}

	// 3. Validate missing argument
	valCmdEmpty, err := parseScenarioValidateCmd(parent, []string{})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmdEmpty.fsys = fsys
	if err := valCmdEmpty.Run(); err == nil || !strings.Contains(err.Error(), "scenario path required") {
		t.Fatalf("expected scenario path required error, got: %v", err)
	}

	// 4. Validate non-existent path in injected FS
	valCmdNotFound, err := parseScenarioValidateCmd(parent, []string{"scenarios/non-existent"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmdNotFound.fsys = fsys
	if err := valCmdNotFound.Run(); err == nil {
		t.Fatal("expected error for non-existent path in FS, got nil")
	}

	// 5. Validate directory missing scenario.txtar in injected FS
	valCmdNoTxtar, err := parseScenarioValidateCmd(parent, []string{"scenarios/no-txtar"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmdNoTxtar.fsys = fsys
	if err := valCmdNoTxtar.Run(); err == nil || !strings.Contains(err.Error(), "missing scenario.txtar") {
		t.Fatalf("expected missing scenario.txtar error, got: %v", err)
	}

	// 6. Validate scenario with missing asset in injected FS
	valCmdMissingAsset, err := parseScenarioValidateCmd(parent, []string{"scenarios/missing-asset/scenario.txtar"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmdMissingAsset.fsys = fsys
	if err := valCmdMissingAsset.Run(); err == nil || !strings.Contains(err.Error(), "asset not found") {
		t.Fatalf("expected asset not found error, got: %v", err)
	}
}

func TestScenarioApplyCmd_InjectedFS(t *testing.T) {
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

	fsys := fstest.MapFS{
		"scenarios/valid/scenario.txtar": &fstest.MapFile{
			Data: []byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: apply-cli-test

-- 01-user.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-pass
At: 2026-08-01T09:00:00Z

-- 02-enable.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:01:00Z
`),
		},
	}

	// Alice creation & enable expectations
	mock.ExpectExec("(?s).*SystemInsertUser.*").
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec("(?s).*InsertUserEmail.*").
		WithArgs(int32(10), "alice@example.test", sql.NullTime{}, sql.NullString{}, nil, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*InsertPassword.*").
		WithArgs(int32(10), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("(?s).*SystemCreateUserRole.*").
		WithArgs(int32(10), "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"apply"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	applyCmd, err := parseScenarioApplyCmd(parent, []string{"scenarios/valid/scenario.txtar"})
	if err != nil {
		t.Fatalf("parseScenarioApplyCmd: %v", err)
	}
	applyCmd.fsys = fsys
	applyCmd.querier = db.New(conn)

	if err := applyCmd.Run(); err != nil {
		t.Fatalf("applyCmd.Run: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestScenarioValidateCmd_ExampleFixtureInMapFS(t *testing.T) {
	fsys := fstest.MapFS{
		"scenarios/100-private-forum/scenario.txtar": &fstest.MapFile{
			Data: []byte(`# Human-readable description.

-- scenario.meta --
Format: goa4web-scenario/v1
Name: private-forum
Description: Staff private forum setup and welcome thread

-- 010-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-test
At: 2026-08-01T09:00:00+10:00

-- 020-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:02:00+10:00

-- 030-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-test
At: 2026-08-01T09:03:00+10:00

-- 040-enable-bob.event --
Op: user.enable
Actor: admin
User: bob
At: 2026-08-01T09:04:00+10:00

-- 050-private-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
Description: Private discussion for Alice and Bob
At: 2026-08-01T09:10:00+10:00
`),
		},
	}

	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"validate"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	valCmd, err := parseScenarioValidateCmd(parent, []string{"scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmd.fsys = fsys

	if err := valCmd.Run(); err != nil {
		t.Fatalf("expected example fixture validation success, got: %v", err)
	}
}

func TestScenarioValidateCmd_CommittedScenario(t *testing.T) {
	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"validate"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	valCmd, err := parseScenarioValidateCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	valCmd.fsys = scenarios.FS

	if err := valCmd.Run(); err != nil {
		t.Fatalf("expected committed scenario validation success, got: %v", err)
	}
}

func TestScenarioCmd_ServeSubcommandDispatch(t *testing.T) {
	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"serve"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmd == nil {
		t.Fatal("expected non-nil scenarioServeCmd")
	}
}
