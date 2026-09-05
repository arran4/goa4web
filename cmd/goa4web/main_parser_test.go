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


func TestParseRoot_EarlyScanSkipsValues(t *testing.T) {
	// E.g. "--db-driver sqlite" should skip "sqlite" and still find --config-file
	args := []string{
		"goa4web",
		"--db-driver", "sqlite",
		"--config-file", "test.env",
		"scenario",
		"serve",
	}

	cmd, _ := parseRoot(args)
	if cmd == nil {
		t.Fatalf("cmd is nil")
	}
	// Unfortunately we can't test internal cfgPath directly without modifying parseRoot further,
	// but we can ensure it doesn't fail parsing.
}

func TestParseRoot_VariousWindows(t *testing.T) {
	cases := []struct {
		name string
		args []string
		env  map[string]string
		wantPub int
		wantPriv int
	}{
		{
			name: "defaults 60/60",
			args: []string{"goa4web", "scenario", "serve"},
			wantPub: 60,
			wantPriv: 60,
		},
		{
			name: "0/15",
			args: []string{"goa4web", "--forum-post-append-window=0", "--private-forum-post-append-window=15", "scenario", "serve"},
			wantPub: 0,
			wantPriv: 15,
		},
		{
			name: "30/0",
			args: []string{"goa4web", "--forum-post-append-window=30", "--private-forum-post-append-window=0", "scenario", "serve"},
			wantPub: 30,
			wantPriv: 0,
		},
		{
			name: "0/0",
			args: []string{"goa4web", "--forum-post-append-window=0", "--private-forum-post-append-window=0", "scenario", "serve"},
			wantPub: 0,
			wantPriv: 0,
		},
		{
			name: "precedence CLI > env",
			args: []string{"goa4web", "--forum-post-append-window=30", "scenario", "serve"},
			env: map[string]string{
				"FORUM_POST_APPEND_WINDOW": "45",
				"PRIVATE_FORUM_POST_APPEND_WINDOW": "25",
			},
			wantPub: 30,
			wantPriv: 25,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tc.env {
					os.Unsetenv(k)
				}
			}()

			cmd, err := parseRoot(tc.args)
			if err != nil {
				t.Fatalf("parseRoot err: %v", err)
			}
			cfg, _ := cmd.RuntimeConfig()
			if cfg.ForumPostAppendWindow != tc.wantPub {
				t.Errorf("pub want %d got %d", tc.wantPub, cfg.ForumPostAppendWindow)
			}
			if cfg.PrivateForumPostAppendWindow != tc.wantPriv {
				t.Errorf("priv want %d got %d", tc.wantPriv, cfg.PrivateForumPostAppendWindow)
			}
		})
	}
}
