package main

import (
	"flag"
	"strings"
	"testing"
	"testing/fstest"
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
At: 2026-08-01T09:00:00Z

-- 02-forum.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Title: Forum 1
At: 2026-08-01T09:02:00Z

-- 03-post.event --
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
At: 2026-08-01T09:00:00+10:00

-- 020-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:02:00+10:00

-- 030-private-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Title: Staff Room
At: 2026-08-01T09:05:00+10:00

-- 040-welcome-post.event --
Op: forum.post
Ref: welcome
Actor: alice
Forum: staff-room
At: 2026-08-01T09:10:00+10:00
Attachment: assets/welcome.jpg

Welcome to the staff forum.
`),
		},
		"scenarios/100-private-forum/assets/welcome.jpg": &fstest.MapFile{
			Data: []byte("mock-image-data"),
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
