//go:build sqlite || sqlite3

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
	"github.com/arran4/goa4web/migrations"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func captureCLIStdout(t *testing.T, fn func()) string {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
	}()

	var buf bytes.Buffer
	copyDone := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&buf, r)
		copyDone <- copyErr
	}()

	fn()
	require.NoError(t, w.Close())
	require.NoError(t, <-copyDone)
	require.NoError(t, r.Close())
	return buf.String()
}

func TestScenarioLegacyReplyEditsCommitSearchAndObservedLinks(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewConfiguredService(q)

	aliceID, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)

	for _, perm := range []struct {
		section, item, action string
	}{
		{"forum", "", "view"},
		{"forum", "", "edit"},
		{"blogs", "post", "reply"},
		{"news", "post", "reply"},
		{"writing", "article", "reply"},
	} {
		require.NoError(t, env.cd.GrantUser(aliceID, perm.section, perm.item, perm.action))
	}

	app := common.NewCoreData(ctx, q, env.server.Config, common.WithJobService(queue)).ForUser(aliceID)

	// Seed a blog post/thread, news post/thread, and writing/thread with initial comments authored by alice.
	_, err := env.db.ExecContext(ctx, `
		INSERT INTO forumtopic (idforumtopic, forumcategory_idforumcategory, title, handler)
			VALUES (901, 1, 'Blog topic', 'blogs'),
			       (902, 1, 'News topic', 'news'),
			       (903, 1, 'Writing topic', 'writing');
		INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic)
			VALUES (901, 901), (902, 902), (903, 903);
		INSERT INTO blogs (idblogs, users_idusers, forumthread_id, blog, written)
			VALUES (901, ?, 901, 'Blog post', CURRENT_TIMESTAMP);
		INSERT INTO site_news (idsiteNews, language_id, news, users_idusers, forumthread_id)
			VALUES (902, 1, 'News post', ?, 902);
		INSERT INTO writing_category (idwritingcategory, title)
			VALUES (903, 'Legacy Writing Cat');
		INSERT INTO writing (idwriting, users_idusers, forumthread_id, writing_category_id, title, writing, abstract, private, published)
			VALUES (903, ?, 903, 903, 'Legacy article', 'Body', 'Abs', 0, CURRENT_TIMESTAMP);
		INSERT INTO comments (idcomments, forumthread_id, users_idusers, language_id, written, text)
			VALUES (9101, 901, ?, 1, CURRENT_TIMESTAMP, 'orig blog reply'),
			       (9102, 902, ?, 1, CURRENT_TIMESTAMP, 'orig news reply'),
			       (9103, 903, ?, 1, CURRENT_TIMESTAMP, 'orig writing reply'),
			       (9104, 903, ?, 1, CURRENT_TIMESTAMP, 'orig article comment');
	`, aliceID, aliceID, aliceID, aliceID, aliceID, aliceID, aliceID)
	require.NoError(t, err)

	readCommentText := func(id int32) string {
		var txt string
		require.NoError(t, env.db.QueryRowContext(ctx, `SELECT text FROM comments WHERE idcomments = ?`, id).Scan(&txt))
		return txt
	}

	assertEditDurable := func(name string, commentID int32, linkURL string, origText string, runEdit func(string) error) {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(func() {
				_, _ = env.db.ExecContext(ctx, `DROP TRIGGER IF EXISTS reject_legacy_edit_job`)
			})
			_, err := env.db.ExecContext(ctx, `DROP TRIGGER IF EXISTS reject_legacy_edit_job`)
			require.NoError(t, err)
			_, err = env.db.ExecContext(ctx, fmt.Sprintf(`CREATE TRIGGER reject_legacy_edit_job BEFORE INSERT ON durable_jobs
				WHEN NEW.job_type = 'search_index' AND NEW.target_id = %d
				BEGIN SELECT RAISE(ABORT, 'injected legacy edit failure'); END`, commentID))
			require.NoError(t, err)

			newText := "updated " + name + " [link=" + linkURL + "]link[/link]"
			err = runEdit(newText)
			require.ErrorContains(t, err, "injected legacy edit failure")
			require.Equal(t, origText, readCommentText(commentID))

			_, err = env.db.ExecContext(ctx, `DROP TRIGGER IF EXISTS reject_legacy_edit_job`)
			require.NoError(t, err)

			require.NoError(t, runEdit(newText))
			savedText := readCommentText(commentID)
			require.NotEqual(t, origText, savedText)
			require.Contains(t, savedText, linkURL)

			dirty, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
				TargetType: searchindex.TypeComment, TargetID: commentID,
			})
			require.NoError(t, err)
			require.GreaterOrEqual(t, dirty.Revision, int64(1))

			var linkJobCount int
			require.NoError(t, env.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM durable_jobs WHERE job_type = 'external_link' AND target_key = ?`, linkURL).Scan(&linkJobCount))
			require.Equal(t, 1, linkJobCount)
		})
	}

	assertEditDurable("blog", 9101, "https://example.test/edit-blog", "orig blog reply", func(text string) error {
		return app.UpdateBlogReply(9101, aliceID, 1, text)
	})
	assertEditDurable("news", 9102, "https://example.test/edit-news", "orig news reply", func(text string) error {
		_, err := app.UpdateNewsReply(9102, aliceID, 1, text)
		return err
	})
	assertEditDurable("writing", 9103, "https://example.test/edit-writing", "orig writing reply", func(text string) error {
		_, err := app.UpdateWritingReply(9103, 1, text)
		return err
	})
	assertEditDurable("article", 9104, "https://example.test/edit-article", "orig article comment", func(text string) error {
		return app.UpdateArticleComment(9104, 1, text)
	})
}

func TestCLIJobsShowEnqueueAndDeactivationCommands(t *testing.T) {
	ctx := t.Context()
	dbPath := filepath.Join(t.TempDir(), "cli_jobs.sqlite")
	sdb, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	sdb.SetMaxOpenConns(1)
	require.NoError(t, dbstart.Apply(ctx, sdb, migrations.FS, false, "sqlite"))
	_, err = sdb.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (1, 'cliuser');
		INSERT INTO comments (idcomments, forumthread_id, users_idusers, language_id, written, text)
			VALUES (55, 1, 1, 1, CURRENT_TIMESTAMP, 'cli deactivation searchable phrase');
	`)
	require.NoError(t, err)
	require.NoError(t, sdb.Close())

	runCLI := func(args ...string) string {
		t.Helper()
		full := append([]string{"goa4web", "--db-driver", "sqlite3", "--db-conn", dbPath}, args...)
		root, err := parseRoot(full)
		require.NoError(t, err)
		defer root.Close()
		var runErr error
		out := captureCLIStdout(t, func() {
			runErr = root.Run()
		})
		require.NoError(t, runErr)
		return out
	}

	// 1. jobs enqueue-url creates external_links row and durable job, then reuses it.
	// targetURL is an external URL exercised by CLI enqueue-url and show tests.
	const targetURL = "https://example.test/cli-unused"
	out1 := runCLI("jobs", "enqueue-url", targetURL)
	require.Contains(t, out1, "Enqueued external_link job for "+targetURL)

	out2 := runCLI("jobs", "enqueue-url", targetURL)
	require.Contains(t, out2, "Coalesced active external_link job for "+targetURL)

	// 2. jobs list includes NEXT RUN column and key.
	listOut := runCLI("jobs", "list")
	require.Contains(t, listOut, "NEXT RUN")
	require.Contains(t, listOut, targetURL)

	// Attach a cooldown record and verify jobs show & show --json.
	sdb2, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	defer func() { _ = sdb2.Close() }()
	var linkID int64
	require.NoError(t, sdb2.QueryRowContext(ctx, `SELECT id FROM external_links WHERE url = ?`, targetURL).Scan(&linkID))
	require.Positive(t, linkID)
	var linkCount int
	require.NoError(t, sdb2.QueryRowContext(ctx, `SELECT COUNT(*) FROM external_links WHERE url = ?`, targetURL).Scan(&linkCount))
	require.Equal(t, 1, linkCount)

	_, err = sdb2.ExecContext(ctx, `INSERT INTO durable_job_cooldowns (dedupe_key, cooldown_until, last_error) VALUES (?, '2030-01-01 00:00:00', 'HTTP 503')`,
		jobs.DedupeKey(jobs.JobTypeExternalLink, targetURL))
	require.NoError(t, err)

	showOut := runCLI("jobs", "show", "1")
	require.Contains(t, showOut, "NextRun:")
	require.Contains(t, showOut, "CooldownUntil:")
	require.Contains(t, showOut, "2030-01-01T00:00:00Z")

	showJSON := runCLI("jobs", "show", "--json", "1")
	var details jobs.JobDetails
	require.NoError(t, json.Unmarshal([]byte(showJSON), &details))
	require.NotNil(t, details.Job)
	require.Equal(t, linkID, details.Job.TargetID)
	require.NotNil(t, details.Cooldown)

	// 3. comment deactivate & comment activate commit durable search_dirty_targets + jobs.
	_ = runCLI("comment", "deactivate", "--id", "55")
	q := db.NewForDriver(sdb2, "sqlite3")
	dirty, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
		TargetType: searchindex.TypeComment, TargetID: 55,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, dirty.Revision)

	_ = runCLI("comment", "activate", "--id", "55")
	dirty, err = q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
		TargetType: searchindex.TypeComment, TargetID: 55,
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, dirty.Revision)
}
