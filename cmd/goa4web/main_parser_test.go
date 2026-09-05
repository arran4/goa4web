package main

import (
	"os"
	"testing"
)

func TestParseRoot_AppendWindows(t *testing.T) {
	args := []string{
		"goa4web",
		"--forum-post-append-window=30",
		"--private-forum-post-append-window=15",
		"scenario",
		"serve",
		"testdata",
	}

	cmd, err := parseRoot(args)
	if err != nil {
		t.Fatalf("parseRoot err: %v", err)
	}
	cfg, _ := cmd.RuntimeConfig()

	if cfg.ForumPostAppendWindow != 30 {
		t.Errorf("expected ForumPostAppendWindow = 30, got %d", cfg.ForumPostAppendWindow)
	}

	if cfg.PrivateForumPostAppendWindow != 15 {
		t.Errorf("expected PrivateForumPostAppendWindow = 15, got %d", cfg.PrivateForumPostAppendWindow)
	}
}

func TestParseRoot_EnvironmentAppendWindows(t *testing.T) {
	os.Setenv("FORUM_POST_APPEND_WINDOW", "45")
	os.Setenv("PRIVATE_FORUM_POST_APPEND_WINDOW", "25")
	defer func() {
		os.Unsetenv("FORUM_POST_APPEND_WINDOW")
		os.Unsetenv("PRIVATE_FORUM_POST_APPEND_WINDOW")
	}()

	args := []string{
		"goa4web",
		"scenario",
		"serve",
		"testdata",
	}

	cmd, err := parseRoot(args)
	if err != nil {
		t.Fatalf("parseRoot err: %v", err)
	}
	cfg, _ := cmd.RuntimeConfig()

	if cfg.ForumPostAppendWindow != 45 {
		t.Errorf("expected ForumPostAppendWindow = 45, got %d", cfg.ForumPostAppendWindow)
	}

	if cfg.PrivateForumPostAppendWindow != 25 {
		t.Errorf("expected PrivateForumPostAppendWindow = 25, got %d", cfg.PrivateForumPostAppendWindow)
	}
}
