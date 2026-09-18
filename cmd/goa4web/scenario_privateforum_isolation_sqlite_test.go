//go:build sqlite || sqlite3

package main

import (
	"context"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/require"
)

func TestPrivateForumIsolation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	require.NoError(t, err)
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	require.NoError(t, err)
	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	require.NoError(t, err)
	serveCmd.fsys = scenarios.FS

	srv, sqlDB, cleanup, err := serveCmd.Bootstrap(ctx)
	require.NoError(t, err)
	defer cleanup()

	cd := common.NewCoreData(ctx, db.NewForDriver(sqlDB, "sqlite3"), srv.Config)

	// get alice id
	var aliceID int32
	err = sqlDB.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
	require.NoError(t, err)

	var projectRoomTopicID int32
	err = sqlDB.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Project Room'").Scan(&projectRoomTopicID)
	require.NoError(t, err)

	var projectRoomThreadID int32
	err = sqlDB.QueryRowContext(ctx, "SELECT forumthread_id FROM comments WHERE text = 'Project Room kickoff for Carol and Dave.'").Scan(&projectRoomThreadID)
	require.NoError(t, err)

	aliceCD := cd.ForUser(aliceID)
	err = aliceCD.ReadForumThread(ctx, common.ReadForumThreadParams{
		ActorID:  aliceID,
		ThreadID: projectRoomThreadID,
	})
	require.Error(t, err, "Expected error when Alice tries to mark Project Room read")

	// Ensure no read marker was created
	marker, err := aliceCD.ThreadReadMarker(projectRoomThreadID)
	require.NoError(t, err)
	require.Equal(t, int32(0), marker, "Expected Alice to have no read marker for Project Room")

	err = aliceCD.SubscribeForum(ctx, common.SubscribeForumParams{
		ActorID: aliceID,
		TopicID: projectRoomTopicID,
	})
	require.Error(t, err, "Expected error when Alice tries to subscribe to Project Room topic")

	// Ensure no topic subscription was created
	var count int
	err = sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscriptions WHERE users_idusers = ? AND pattern LIKE '%topic/' || ? || '/%'", aliceID, projectRoomTopicID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "Expected Alice to have no subscriptions to Project Room topic")

	err = aliceCD.UnsubscribeForum(ctx, common.SubscribeForumParams{
		ActorID: aliceID,
		TopicID: projectRoomTopicID,
	})
	require.Error(t, err, "Expected error when Alice tries to unsubscribe from Project Room topic")
}
