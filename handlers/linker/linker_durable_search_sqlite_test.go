//go:build sqlite || sqlite3

package linker

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
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestApprovedLinkCommitsWithSearchIntent(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite"))

	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (50, 'link moderator');
		INSERT INTO linker_category (id, title) VALUES (50, 'Searchable resources');
		INSERT INTO linker_queue (id, submitter_id, category_id, title, url, description)
			VALUES (50, 50, 50, 'Queued reference', 'https://example.test/linker-reference', 'Article index');
		INSERT INTO grants (user_id, section, item, rule_type, action)
			VALUES (50, 'linker', 'link', 'allow', 'view');
		CREATE TRIGGER reject_linker_search BEFORE INSERT ON durable_jobs
			WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'linker'
			BEGIN SELECT RAISE(ABORT, 'injected linker search failure'); END;
	`)
	require.NoError(t, err)
	q := db.NewForDriver(conn, "sqlite")
	queue := jobs.NewService(q)
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithJobService(queue)).ForUser(50)
	approve := func(query string) any {
		req := httptest.NewRequest("POST", "/admin/queue?"+query, nil)
		req = req.WithContext(context.WithValue(ctx, consts.KeyCoreData, cd))
		return AdminApproveTask.Action(httptest.NewRecorder(), req)
	}

	result := approve("qid=50")
	var count int
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM linker WHERE title = 'Queued reference'").Scan(&count))
	require.Zero(t, count, "a failed search job insert must roll back approval")
	approvalErr, _ := result.(error)
	require.ErrorContains(t, approvalErr, "injected linker search failure")

	_, err = conn.ExecContext(ctx, "DROP TRIGGER reject_linker_search")
	require.NoError(t, err)
	require.Nil(t, approve("qid=50"))
	claim, err := queue.ClaimNext(ctx, "standalone-search", jobs.JobTypeSearchIndex)
	require.NoError(t, err)
	require.NotNil(t, claim)
	require.Equal(t, searchindex.TypeLinker, claim.Job.TargetType)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM linker WHERE id = ? AND listed IS NOT NULL", claim.Job.TargetID).Scan(&count))
	require.Equal(t, 1, count)

	_, err = conn.ExecContext(ctx, `INSERT INTO linker_queue (id, submitter_id, category_id, title, url)
		VALUES (51, 50, 50, 'First bulk link', 'https://example.test/first'),
		       (52, 50, 50, 'Second bulk link', 'https://example.test/second')`)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/admin/queue?qid=51&qid=52", nil)
	req = req.WithContext(context.WithValue(ctx, consts.KeyCoreData, cd))
	require.Nil(t, AdminBulkApproveTask.Action(httptest.NewRecorder(), req))
	var bulkTitles []string
	for range 2 {
		bulkClaim, err := queue.ClaimNext(ctx, "standalone-search", jobs.JobTypeSearchIndex)
		require.NoError(t, err)
		require.NotNil(t, bulkClaim)
		require.Equal(t, searchindex.TypeLinker, bulkClaim.Job.TargetType)
		var title string
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT title FROM linker WHERE id = ? AND listed IS NOT NULL", bulkClaim.Job.TargetID).Scan(&title))
		bulkTitles = append(bulkTitles, title)
	}
	require.ElementsMatch(t, []string{"First bulk link", "Second bulk link"}, bulkTitles)

	_, err = conn.ExecContext(ctx, `
		DELETE FROM grants WHERE user_id = 50 AND section = 'linker';
		INSERT INTO linker_queue (id, submitter_id, category_id, title, url)
			VALUES (53, 50, 50, 'Approval without viewing', 'https://example.test/approval');
	`)
	require.NoError(t, err)
	require.Nil(t, approve("qid=53"), "a successful approval survives an unavailable post-commit read")
	postCommit, err := queue.ClaimNext(ctx, "standalone-search", jobs.JobTypeSearchIndex)
	require.NoError(t, err)
	require.NotNil(t, postCommit)
	require.Equal(t, searchindex.TypeLinker, postCommit.Job.TargetType)
}

func TestAdminUpdateLinkerItemAndEditReplyCommitDurableSearch(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite"))

	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (50, 'linker editor');
		INSERT INTO linker_category (id, title) VALUES (50, 'Resources');
		INSERT INTO forumtopic (idforumtopic, forumcategory_idforumcategory, title, handler)
			VALUES (50, 1, 'A LINKER TOPIC', 'linker');
		INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic) VALUES (50, 50);
		INSERT INTO linker (id, language_id, author_id, category_id, thread_id, title, url, description, listed)
			VALUES (50, 1, 50, 50, 50, 'Original link title', 'https://example.test/orig', 'Original description', CURRENT_TIMESTAMP);
		INSERT INTO comments (idcomments, forumthread_id, users_idusers, language_id, written, text)
			VALUES (50, 50, 50, 1, CURRENT_TIMESTAMP, 'initial linker reply');
		INSERT INTO grants (user_id, section, item, item_id, rule_type, action, active)
			VALUES (50, 'forum', NULL, NULL, 'allow', 'edit', 1),
			       (50, 'forum', 'topic', 50, 'allow', 'view', 1);
	`)
	require.NoError(t, err)

	q := db.NewForDriver(conn, "sqlite")
	queue := jobs.NewConfiguredService(q)
	session := sessions.NewSession(nil, "linker-test")
	session.Values["UID"] = int32(50)
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserID(50), common.WithJobService(queue))
	cd.SetSession(session)

	// 1. AdminEditLinkTask rolls back on durable_jobs failure and updates search on retry.
	_, err = conn.ExecContext(ctx, `CREATE TRIGGER reject_linker_edit_search BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'linker'
		BEGIN SELECT RAISE(ABORT, 'injected linker edit failure'); END`)
	require.NoError(t, err)

	runEditLink := func(title, desc string) any {
		form := url.Values{
			"title":    {title},
			"URL":      {"https://example.test/updated"},
			"desc":     {desc},
			"category": {"50"},
			"language": {"1"},
		}
		req := httptest.NewRequest(http.MethodPost, "/admin/linker/links/link/50", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"link": "50"})
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		return AdminEditLinkTask.Action(httptest.NewRecorder(), req)
	}

	res := runEditLink("Modified link title", "Modified unique token")
	editErr, ok := res.(error)
	require.True(t, ok)
	require.ErrorContains(t, editErr, "injected linker edit failure")
	var curTitle string
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT title FROM linker WHERE id = 50`).Scan(&curTitle))
	require.Equal(t, "Original link title", curTitle)

	_, err = conn.ExecContext(ctx, `DROP TRIGGER reject_linker_edit_search`)
	require.NoError(t, err)
	require.NotNil(t, runEditLink("Modified link title", "FreshlyIndexedLinkerWord"))

	// 2. Linker EditReplyTask rolls back on failure, then records durable search + observed link work.
	_, err = conn.ExecContext(ctx, `CREATE TRIGGER reject_linker_reply_search BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'comment'
		BEGIN SELECT RAISE(ABORT, 'injected linker reply failure'); END`)
	require.NoError(t, err)

	runEditReply := func(text string) any {
		form := url.Values{"language": {"1"}, "replytext": {text}}
		req := httptest.NewRequest(http.MethodPost, "/linker/comments/50/edit/50", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"link": "50", "comment": "50"})
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		return commentEditAction.Action(httptest.NewRecorder(), req)
	}

	replyRes := runEditReply("edited [a=https://example.test/linker-reply-link]LinkerReplyKeyword[/a]")
	replyErr, ok := replyRes.(error)
	require.True(t, ok)
	require.ErrorContains(t, replyErr, "injected linker reply failure")
	var curReply string
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT text FROM comments WHERE idcomments = 50`).Scan(&curReply))
	require.Equal(t, "initial linker reply", curReply)

	_, err = conn.ExecContext(ctx, `DROP TRIGGER reject_linker_reply_search`)
	require.NoError(t, err)
	_ = runEditReply("edited [a=https://example.test/linker-reply-link]LinkerReplyKeyword[/a]")
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT text FROM comments WHERE idcomments = 50`).Scan(&curReply))
	require.Contains(t, curReply, "LinkerReplyKeyword")

	linkJob, err := queue.ClaimNext(ctx, "link-worker", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, linkJob)
	require.Equal(t, "https://example.test/linker-reply-link", linkJob.Job.TargetKey)

	// Run durable search worker and confirm both linker and comment search indexes updated.
	wCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		searchworker.RunDurableWorker(wCtx, queue, q, "linker-search-worker")
	}()
	require.Eventually(t, func() bool {
		lState, err1 := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{TargetType: searchindex.TypeLinker, TargetID: 50})
		cState, err2 := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{TargetType: searchindex.TypeComment, TargetID: 50})
		return err1 == nil && err2 == nil && lState.IndexedRevision == 1 && cState.IndexedRevision == 1
	}, 4*time.Second, 15*time.Millisecond)
	cancel()
	<-done

	var wordCount int
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM linker_search ls
		JOIN searchwordlist sw ON sw.idsearchwordlist = ls.searchwordlist_idsearchwordlist
		WHERE ls.linker_id = 50 AND sw.word = 'freshlyindexedlinkerword'`).Scan(&wordCount))
	require.Equal(t, 1, wordCount)

	require.NoError(t, conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments_search cs
		JOIN searchwordlist sw ON sw.idsearchwordlist = cs.searchwordlist_idsearchwordlist
		WHERE cs.comment_id = 50 AND sw.word = 'linkerreplykeyword'`).Scan(&wordCount))
	require.Equal(t, 1, wordCount)
}
