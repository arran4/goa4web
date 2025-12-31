package config

import (
	"os"
	"os/exec"
	"testing"
)

func TestDBConfigConflicts(t *testing.T) {
	// Since normalizeRuntimeConfig uses log.Fatalf, we need to test it in a subprocess
	if os.Getenv("TEST_DB_CONFIG_CONFLICT") == "1" {
		cfg := &RuntimeConfig{
			DBConn: "user:pass@tcp(127.0.0.1:3306)/dbname",
			DBUser: "otheruser",
		}
		normalizeRuntimeConfig(cfg)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDBConfigConflicts")
	cmd.Env = append(os.Environ(), "TEST_DB_CONFIG_CONFLICT=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("process ran with err %v, want exit status 1", err)
	}
}

func TestDBConfigConflictsPass(t *testing.T) {
	if os.Getenv("TEST_DB_CONFIG_CONFLICT") == "1" {
		cfg := &RuntimeConfig{
			DBConn: "user:pass@tcp(127.0.0.1:3306)/dbname",
			DBPass: "otherpass",
		}
		normalizeRuntimeConfig(cfg)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDBConfigConflictsPass")
	cmd.Env = append(os.Environ(), "TEST_DB_CONFIG_CONFLICT=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("process ran with err %v, want exit status 1", err)
	}
}

func TestDBConfigConflictsHost(t *testing.T) {
	if os.Getenv("TEST_DB_CONFIG_CONFLICT") == "1" {
		cfg := &RuntimeConfig{
			DBConn: "user:pass@tcp(127.0.0.1:3306)/dbname",
			DBHost: "otherhost",
		}
		normalizeRuntimeConfig(cfg)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDBConfigConflictsHost")
	cmd.Env = append(os.Environ(), "TEST_DB_CONFIG_CONFLICT=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("process ran with err %v, want exit status 1", err)
	}
}

func TestDBConfigConflictsPort(t *testing.T) {
	if os.Getenv("TEST_DB_CONFIG_CONFLICT") == "1" {
		cfg := &RuntimeConfig{
			DBConn: "user:pass@tcp(127.0.0.1:3306)/dbname",
			DBPort: "3307",
		}
		normalizeRuntimeConfig(cfg)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDBConfigConflictsPort")
	cmd.Env = append(os.Environ(), "TEST_DB_CONFIG_CONFLICT=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("process ran with err %v, want exit status 1", err)
	}
}

func TestDBConfigReconstruction(t *testing.T) {
	cfg := &RuntimeConfig{
		DBUser: "user",
		DBPass: "pass",
		DBHost: "127.0.0.1",
		DBPort: "3306",
		DBName: "dbname",
	}
	normalizeRuntimeConfig(cfg)
	expected := "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true"
	if cfg.DBConn != expected {
		t.Errorf("expected %s, got %s", expected, cfg.DBConn)
	}
}
