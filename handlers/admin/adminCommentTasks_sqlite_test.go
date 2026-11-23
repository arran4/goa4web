//go:build sqlite || sqlite3

package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
	"github.com/arran4/goa4web/migrations"
	"github.com/arran4/goa4web/workers/searchworker"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAdminCommentTasksCommitDurableSearchAndLinkWork(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite"))

	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (1, 'admin'), (2, 'author');
		INSERT INTO forumtopic (idforumtopic, forumcategory_idforumcategory, title, handler)
			VALUES (1, 1, 'Public topic', 'forum');
		INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic) VALUES (1, 1);
		INSERT INTO grants (user_id, section, item, item_id, rule_type, action, active)
			VALUES (1, 'forum', 'topic', 1, 'allow', 'view', 1);
		INSERT INTO comments (idcomments, forumthread_id, users_idusers, language_id, written, text)
			VALUES (10, 1, 2, 0, CURRENT_TIMESTAMP, 'initial admin searchable text');
	`)
	require.NoError(t, err)

	q := db.NewForDriver(conn, "sqlite")
	svc := jobs.NewConfiguredService(q)
	require.NoError(t, searchworker.Index(ctx, q, searchworker.IndexEventData{
		Type: searchworker.TypeComment, ID: 10, Text: "initial admin searchable text",
	}))

	readCommentWords := func() []string {
		t.Helper()
		rows, err := conn.QueryContext(ctx, `SELECT sw.word FROM searchwordlist sw
			JOIN comments_search cs ON cs.searchwordlist_idsearchwordlist = sw.idsearchwordlist
			WHERE cs.comment_id = 10`)
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()
		var words []string
		for rows.Next() {
			var w string
			require.NoError(t, rows.Scan(&w))
			words = append(words, w)
		}
		require.NoError(t, rows.Err())
		return words
	}

	readCommentText := func() string {
		t.Helper()
		var text string
		require.NoError(t, conn.QueryRowContext(ctx, `SELECT COALESCE(text, '') FROM comments WHERE idcomments = 10`).Scan(&text))
		return text
	}

	runDurableSearchPass := func(wantRevision int64) {
		t.Helper()
		wCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			searchworker.RunDurableWorker(wCtx, svc, q, "admin-search-worker")
		}()
		require.Eventually(t, func() bool {
			state, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
				TargetType: searchindex.TypeComment, TargetID: 10,
			})
			return err == nil && state.Revision == wantRevision && state.IndexedRevision == wantRevision
		}, 4*time.Second, 15*time.Millisecond)
		cancel()
		<-done
	}

	makeReq := func(form url.Values) *http.Request {
		var bodyStr string
		if form != nil {
			bodyStr = form.Encode()
		}
		req := httptest.NewRequest(http.MethodPost, "/admin/comment/10", strings.NewReader(bodyStr))
		if form != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		req = mux.SetURLVars(req, map[string]string{"comment": "10"})
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserID(1), common.WithJobService(svc))
		cd.LoadSelectionsFromRequest(req)
		return req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	}

	// 1. Trigger failure rolls back editCommentTask cleanly.
	_, err = conn.ExecContext(ctx, `CREATE TRIGGER reject_admin_comment_search BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'comment'
		BEGIN SELECT RAISE(ABORT, 'injected admin comment search failure'); END`)
	require.NoError(t, err)

	res := editCommentTask.Action(httptest.NewRecorder(), makeReq(url.Values{
		"replytext": {"edited [a=https://example.test/admin-link]updatedterm[/a]"},
	}))
	actionErr, ok := res.(error)
	require.True(t, ok)
	require.ErrorContains(t, actionErr, "injected admin comment search failure")
	require.Equal(t, "initial admin searchable text", readCommentText())

	_, err = conn.ExecContext(ctx, `DROP TRIGGER reject_admin_comment_search`)
	require.NoError(t, err)

	// 2. Edit commits durable search + external link work, preserves deleted_at IS NULL, and updates search tokens.
	editedText := "edited [a=https://example.test/admin-link]updatedterm[/a]"
	require.Nil(t, editCommentTask.Action(httptest.NewRecorder(), makeReq(url.Values{"replytext": {editedText}})))
	require.Equal(t, editedText, readCommentText())
	var deletedAt sql.NullTime
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT deleted_at FROM comments WHERE idcomments = 10`).Scan(&deletedAt))
	require.False(t, deletedAt.Valid, "editing a comment must preserve deleted_at IS NULL")
	linkClaim, err := svc.ClaimNext(ctx, "link-worker", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, linkClaim)
	require.Equal(t, "https://example.test/admin-link", linkClaim.Job.TargetKey)
	_, err = svc.CompleteFenced(ctx, linkClaim)
	require.NoError(t, err)

	runDurableSearchPass(1)
	words := readCommentWords()
	require.Contains(t, words, "updatedterm")
	require.NotContains(t, words, "initial")

	// 3. Deactivate archives, scrubs, and clears search index tokens.
	require.Nil(t, deactivateCommentTask.Action(httptest.NewRecorder(), makeReq(nil)))
	require.Equal(t, "", readCommentText())
	runDurableSearchPass(2)
	require.Empty(t, readCommentWords())

	// 4. Restore recovers archived text, search index, and link observation.
	_, err = conn.ExecContext(ctx, `UPDATE durable_jobs SET active_dedupe_key = NULL WHERE job_type = 'external_link'`)
	require.NoError(t, err)
	require.Nil(t, restoreCommentTask.Action(httptest.NewRecorder(), makeReq(nil)))
	require.Equal(t, editedText, readCommentText())
	runDurableSearchPass(3)
	require.Contains(t, readCommentWords(), "updatedterm")
	linkClaim2, err := svc.ClaimNext(ctx, "link-worker-2", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, linkClaim2)
	require.Equal(t, "https://example.test/admin-link", linkClaim2.Job.TargetKey)

	// 5. Delete scrubs text and clears search index tokens.
	require.Nil(t, deleteCommentTask.Action(httptest.NewRecorder(), makeReq(nil)))
	require.Equal(t, "", readCommentText())
	runDurableSearchPass(4)
	require.Empty(t, readCommentWords())
}
