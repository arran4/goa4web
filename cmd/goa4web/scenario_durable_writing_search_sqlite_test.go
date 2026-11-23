//go:build sqlite || sqlite3

package main

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
	"github.com/stretchr/testify/require"
)

func TestScenarioWritingCommitsWithSearchIntent(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	alice, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	require.NoError(t, q.AdminInsertWritingCategory(ctx, db.AdminInsertWritingCategoryParams{
		WritingCategoryID: sql.NullInt32{Valid: true},
		Title:             sql.NullString{String: "Search index scenario", Valid: true},
	}))
	var category int32
	err := env.db.QueryRowContext(ctx, `SELECT idwritingcategory FROM writing_category WHERE title = 'Search index scenario'`).Scan(&category)
	require.NoError(t, err)
	require.NoError(t, env.cd.GrantWritingCategory(category, alice, 0, []string{"post", "list", "see", "view"}))
	for _, action := range []string{"post", "view"} {
		require.NoError(t, env.cd.GrantUser(alice, "writing", "article", action))
	}
	queue := jobs.NewService(q)
	app := common.NewCoreData(ctx, q, env.server.Config, common.WithJobService(queue)).ForUser(alice)

	_, err = env.db.ExecContext(ctx, `CREATE TRIGGER reject_writing_search BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'writing'
		BEGIN SELECT RAISE(ABORT, 'injected writing search failure'); END`)
	require.NoError(t, err)
	id, err := app.CreateWriting(category, 1, "Searchable article", "An abstract", "Authoritative prose", false)
	require.ErrorContains(t, err, "injected writing search failure")
	require.Zero(t, id)
	var count int
	err = env.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM writing WHERE title = 'Searchable article'`).Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)

	_, err = env.db.ExecContext(ctx, "DROP TRIGGER reject_writing_search")
	require.NoError(t, err)
	id, err = app.CreateWriting(category, 1, "Searchable article", "An abstract", "Authoritative prose", false)
	require.NoError(t, err)
	require.NotZero(t, id)
	claim, err := queue.ClaimNext(ctx, "standalone-search", jobs.JobTypeSearchIndex)
	require.NoError(t, err)
	require.NotNil(t, claim)
	require.Equal(t, searchindex.TypeWriting, claim.Job.TargetType)
	require.Equal(t, id, claim.Job.TargetID)

	get := func() *db.GetWritingForListerByIDRow {
		article, err := q.GetWritingForListerByID(ctx, db.GetWritingForListerByIDParams{
			Idwriting: int32(id), ListerID: alice, ListerMatchID: sql.NullInt32{Int32: alice, Valid: true},
		})
		require.NoError(t, err)
		return article
	}
	article := get()
	_, err = env.db.ExecContext(ctx, `CREATE TRIGGER reject_writing_search_revision BEFORE UPDATE ON search_dirty_targets
		WHEN NEW.target_type = 'writing' BEGIN SELECT RAISE(ABORT, 'injected writing revision failure'); END`)
	require.NoError(t, err)
	err = app.UpdateWriting(article, article.Title.String, article.Abstract.String, "Revised prose", false, 1)
	require.ErrorContains(t, err, "injected writing revision failure")
	require.Equal(t, "Authoritative prose", get().Writing.String)

	_, err = env.db.ExecContext(ctx, "DROP TRIGGER reject_writing_search_revision")
	require.NoError(t, err)
	require.NoError(t, app.UpdateWriting(article, article.Title.String, article.Abstract.String, "Revised prose", false, 1))
	require.Equal(t, "Revised prose", get().Writing.String)
	active, err := q.SystemGetDurableJobByID(ctx, claim.Job.ID)
	require.NoError(t, err)
	require.EqualValues(t, 2, active.TargetRevision)
	require.True(t, active.NeedsRerun)
}
