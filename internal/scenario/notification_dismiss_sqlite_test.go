//go:build sqlite || sqlite3

package scenario

import (
	"database/sql"
	"io/fs"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/sqlutil"
	"github.com/arran4/goa4web/migrations"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/require"
)

func TestScenarioNotificationDismissalPreservesThreadReadPosition(t *testing.T) {
	for _, tc := range []struct {
		name       string
		actor      string
		appendMode bool
	}{
		{name: "recipient dismisses only the selected reply", actor: "alice"},
		{name: "recipient dismisses the appended activity", actor: "alice", appendMode: true},
		{name: "nonrecipient cannot dismiss the reply", actor: "dave"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			conn, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			conn.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = conn.Close() })
			require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite3"))
			require.NoError(t, sqlutil.RunStatements(ctx, conn, strings.NewReader(string(database.SeedSQLForDriver("sqlite3")))))
			_, err = conn.ExecContext(ctx, "INSERT INTO language (id, nameof) VALUES (1, 'English')")
			require.NoError(t, err)

			fixture, err := fs.Sub(scenarios.FS, "100-private-forum")
			require.NoError(t, err)
			data, err := fs.ReadFile(fixture, "scenario.txtar")
			require.NoError(t, err)
			if tc.actor != "alice" {
				data = []byte(strings.Replace(string(data), "Op: notification.dismiss\nActor: alice", "Op: notification.dismiss\nActor: "+tc.actor, 1))
			}
			sc, err := Parse(data, fixture)
			require.NoError(t, err)
			targetRef := "bob-staff-welcome-reply-2"
			originalRef := ""
			if tc.appendMode {
				// The append query uses the database wall clock. Give two
				// consecutive, unseen replies current timestamps.
				dismiss := sc.Events[len(sc.Events)-1]
				now := time.Now().UTC().Truncate(time.Second)
				originalRef = "bob-unseen-reply"
				targetRef = "bob-appends-unseen-reply"
				replyEvent := func(file, ref, body string, at time.Time) *Event {
					headers := NewHeader()
					headers.Add("Op", "forum.reply")
					headers.Add("Ref", ref)
					headers.Add("Actor", "bob")
					headers.Add("Thread", "staff-welcome")
					headers.Add("At", at.Format(time.RFC3339))
					return &Event{File: file, Op: "forum.reply", At: at, Headers: headers, Body: body}
				}
				dismiss.At = now
				dismiss.Headers.Set("At", now.Format(time.RFC3339))
				dismiss.Headers.Set("Reply", targetRef)
				sc.Events = append(sc.Events[:len(sc.Events)-1],
					replyEvent("638-bob-unseen.event", originalRef, "Here is the plan.", now.Add(-2*time.Minute)),
					replyEvent("639-bob-append.event", targetRef, "One more detail about the plan.", now.Add(-time.Minute)),
					dismiss,
				)
			}
			q := db.NewForDriver(conn, "sqlite3")
			runner := NewRunnerWithQuerier(ctx, q)
			if tc.appendMode {
				runner.coreData.Config.PrivateForumPostAppendWindow = 60
			}
			_, err = runner.Apply(ctx, sc)
			if tc.actor == "alice" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, sql.ErrNoRows)
			}

			alice, ok := runner.Registry().ResolveUser("alice")
			require.True(t, ok)
			threadID, ok := runner.resolveInt32Ref(RefTypeThread, "staff-welcome")
			require.True(t, ok)
			firstReplyID, ok := runner.resolveInt32Ref(RefTypePost, "bob-welcome-reply")
			require.True(t, ok)
			targetReplyID, ok := runner.resolveInt32Ref(RefTypePost, targetRef)
			require.True(t, ok)
			if tc.appendMode {
				originalID, ok := runner.resolveInt32Ref(RefTypePost, originalRef)
				require.True(t, ok)
				require.Equal(t, originalID, targetReplyID)
			}
			activity, ok := runner.Registry().Resolve(refTypeReplyActivity, targetRef)
			require.True(t, ok)
			activityID, ok := activity.(string)
			require.True(t, ok)
			targetKey := notifications.ForumReplyActivityKey(activityID)
			cd := runner.coreData.ForUser(alice)
			watermark, err := cd.ThreadReadMarkerPosition(threadID)
			require.NoError(t, err)
			require.Equal(t, common.CalculateActivityPosition(firstReplyID, 0), watermark)

			notes, err := cd.UserNotifications(alice, 20, 0)
			require.NoError(t, err)
			var target, unrelated, original *db.Notification
			for _, note := range notes {
				if note.ResourceID.Int32 == threadID && note.ActivityKey.String == targetKey {
					target = note
				}
				if note.ResourceID.Valid && note.ResourceID.Int32 != threadID {
					unrelated = note
				}
				if note.ResourceID.Int32 == threadID && note.ActivityPosition.Int64 == common.CalculateActivityPosition(targetReplyID, 0) {
					original = note
				}
			}
			require.NotNil(t, target)
			require.Equal(t, tc.actor == "alice", target.ReadAt.Valid)
			require.NotNil(t, unrelated)
			require.False(t, unrelated.ReadAt.Valid)
			if tc.appendMode {
				require.NotNil(t, original)
				require.False(t, original.ReadAt.Valid)
			}
		})
	}
}
