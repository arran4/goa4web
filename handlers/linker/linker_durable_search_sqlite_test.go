//go:build sqlite || sqlite3

package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
	"github.com/arran4/goa4web/migrations"
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
