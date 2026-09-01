//go:build sqlite || sqlite3

package main

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/testdata/scenarios"
)

func TestScenarioServePrivateForumUsersHaveNavigationCapability(t *testing.T) {
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

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanup()

	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, srv.Config)

	for _, username := range []string{"alice", "bob", "carol", "dave"} {
		t.Run(username, func(t *testing.T) {
			var userID int32
			if err := dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = ?", username).Scan(&userID); err != nil {
				t.Fatalf("query %s: %v", username, err)
			}

			userCD := cd.ForUser(userID)
			if !userCD.HasGrant("privateforum", "topic", "view", 0) {
				t.Fatalf("%s lacks global privateforum/topic/view capability", username)
			}

			items := srv.Nav.IndexItemsWithPermission(func(section, item string) bool {
				return userCD.HasGrant(section, item, "view", 0) || userCD.IsAdmin()
			})
			for _, item := range items {
				if item.Name == "Private" && item.Link == "/private" {
					return
				}
			}
			t.Fatalf("%s cannot see the Private navigation entry", username)
		})
	}
}
