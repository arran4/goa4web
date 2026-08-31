package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestScenarioValidateCmd_ValidFileAndDir(t *testing.T) {
	// Create temporary directory with scenario and asset
	tmpDir := t.TempDir()
	scenarioTxtar := filepath.Join(tmpDir, "scenario.txtar")
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "img.jpg"), []byte("data"), 0644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	content := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: cli-test

-- 01-user.event --
Op: user.create
Ref: alice
Username: alice
At: 2026-08-01T09:00:00Z

-- 02-enable.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:01:00Z
`
	if err := os.WriteFile(scenarioTxtar, []byte(content), 0644); err != nil {
		t.Fatalf("write scenario: %v", err)
	}

	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"validate"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	// 1. Validate by file path
	valCmd, err := parseScenarioValidateCmd(parent, []string{scenarioTxtar})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	if err := valCmd.Run(); err != nil {
		t.Fatalf("expected validation success for file, got: %v", err)
	}

	// 2. Validate by directory path
	valCmdDir, err := parseScenarioValidateCmd(parent, []string{tmpDir})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	if err := valCmdDir.Run(); err != nil {
		t.Fatalf("expected validation success for dir, got: %v", err)
	}

	// 3. Validate missing argument
	valCmdEmpty, err := parseScenarioValidateCmd(parent, []string{})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	if err := valCmdEmpty.Run(); err == nil || !strings.Contains(err.Error(), "scenario path required") {
		t.Fatalf("expected scenario path required error, got: %v", err)
	}

	// 4. Validate non-existent path
	valCmdNotFound, err := parseScenarioValidateCmd(parent, []string{filepath.Join(tmpDir, "non-existent")})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	if err := valCmdNotFound.Run(); err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestScenarioValidateCmd_ExampleFixture(t *testing.T) {
	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"validate"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	fixturePath := filepath.Join("..", "..", "testdata", "scenarios", "100-private-forum")
	valCmd, err := parseScenarioValidateCmd(parent, []string{fixturePath})
	if err != nil {
		t.Fatalf("parseScenarioValidateCmd: %v", err)
	}
	if err := valCmd.Run(); err != nil {
		t.Fatalf("expected example fixture validation success, got: %v", err)
	}
}
