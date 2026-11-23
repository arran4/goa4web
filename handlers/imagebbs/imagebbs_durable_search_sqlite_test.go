//go:build sqlite || sqlite3

package imagebbs

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
	"github.com/arran4/goa4web/internal/upload/local"
	"github.com/arran4/goa4web/migrations"
	"github.com/arran4/goa4web/workers"
	"github.com/arran4/goa4web/workers/searchworker"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestImageBoardUploadCommitsWithSearchIntent(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.ImageMaxBytes = 1024 * 1024
	cfg.ImageUploadProvider = ""
	conn, _, queue, run := imageBoardUploadFixture(t, cfg)
	ctx := t.Context()
	_, err := conn.ExecContext(ctx, `CREATE TRIGGER reject_image_search BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'image'
		BEGIN SELECT RAISE(ABORT, 'injected image search failure'); END;`)
	require.NoError(t, err)

	result := run()
	var count int
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM imagepost WHERE imageboard_idimageboard = 50").Scan(&count))
	require.Zero(t, count, "a failed search job insert must roll back the post")
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM uploaded_images WHERE users_idusers = 50").Scan(&count))
	require.Zero(t, count, "a failed search job insert must roll back the upload record")
	uploadErr, _ := result.(error)
	require.ErrorContains(t, uploadErr, "injected image search failure")

	_, err = conn.ExecContext(ctx, "DROP TRIGGER reject_image_search")
	require.NoError(t, err)
	result = run()
	_, failed := result.(error)
	require.False(t, failed, "upload retry failed: %v", result)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM uploaded_images WHERE users_idusers = 50").Scan(&count))
	require.Equal(t, 1, count, "retry records the upload exactly once")
	claim, err := queue.ClaimNext(ctx, "standalone-search", jobs.JobTypeSearchIndex)
	require.NoError(t, err)
	require.NotNil(t, claim)
	require.Equal(t, searchindex.TypeImage, claim.Job.TargetType)
	var caption string
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT description FROM imagepost WHERE idimagepost = ?", claim.Job.TargetID).Scan(&caption))
	require.Equal(t, "Searchable image caption", caption)
}

func TestHappyPathImageBoardThumbnailRecoversAfterWorkerRestart(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.ImageMaxBytes = 1024 * 1024
	cfg.ImageUploadProvider = "local"
	// Use the real local upload provider to exercise files that survive a worker restart.
	cfg.ImageUploadDir = t.TempDir()
	local.Register()
	conn, q, queue, run := imageBoardUploadFixture(t, cfg)
	ctx := t.Context()
	var thumbnail string
	require.True(t, t.Run("Upload commits before dispatch", func(t *testing.T) {
		require.Nil(t, run())
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT thumbnail FROM imagepost").Scan(&thumbnail))
	}))
	relative, ok := strings.CutPrefix(thumbnail, "/imagebbs/images/")
	require.True(t, ok)
	name := filepath.Join(cfg.ImageUploadDir, relative)
	var jobID int64
	require.True(t, t.Run("A worker is interrupted after claiming the thumbnail", func(t *testing.T) {
		_, err := os.Stat(name)
		require.ErrorIs(t, err, os.ErrNotExist)
		claim, err := queue.ClaimNext(ctx, "interrupted-worker", jobs.JobTypeImageBoardThumbnail)
		require.NoError(t, err)
		require.NotNil(t, claim)
		jobID = claim.Job.ID
	}))
	t.Run("The standalone worker reclaims the job and writes the thumbnail", func(t *testing.T) {
		replacement := jobs.NewService(q)
		replacement.Now = func() time.Time { return time.Now().Add(2 * queue.LeaseDuration) }
		group := workers.StartDurableWorkers(ctx, q, cfg, replacement, "replacement")
		t.Cleanup(func() {
			group.Cancel()
			wait, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			require.NoError(t, group.Wait(wait))
		})
		require.Eventually(t, func() bool {
			job, err := q.SystemGetDurableJobByID(ctx, jobID)
			return err == nil && job.Status == jobs.StatusCompleted && job.Attempts == 2
		}, 5*time.Second, 20*time.Millisecond)
		data, err := os.ReadFile(name)
		require.NoError(t, err)
		img, err := png.Decode(bytes.NewReader(data))
		require.NoError(t, err)
		require.Equal(t, image.Rect(0, 0, 200, 200), img.Bounds())
	})
}

func TestImageBoardUploadRollsBackWhenThumbnailJobCannotBeRecorded(t *testing.T) {
	t.Run("Unhappy Path - thumbnail insert fails", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.ImageMaxBytes = 1024 * 1024
		cfg.ImageUploadProvider = "local"
		// The original is written before the database transaction through the real upload provider.
		cfg.ImageUploadDir = t.TempDir()
		local.Register()
		conn, _, _, run := imageBoardUploadFixture(t, cfg)
		ctx := t.Context()
		_, err := conn.ExecContext(ctx, `CREATE TRIGGER reject_thumbnail BEFORE INSERT ON durable_jobs
			WHEN NEW.job_type = 'imageboard_thumbnail'
			BEGIN SELECT RAISE(ABORT, 'injected thumbnail failure'); END;`)
		require.NoError(t, err)
		result, _ := run().(error)
		require.ErrorContains(t, result, "injected thumbnail failure")
		var count int
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM imagepost").Scan(&count))
		require.Zero(t, count)
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM uploaded_images").Scan(&count))
		require.Zero(t, count)
	})
}

func imageBoardUploadFixture(t *testing.T, cfg *config.RuntimeConfig) (*sql.DB, db.Querier, *jobs.Service, func() any) {
	t.Helper()
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite"))
	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (50, 'image poster');
		INSERT INTO imageboard (idimageboard, imageboard_idimageboard, title) VALUES (50, 0, 'Images');
		INSERT INTO grants (user_id, section, item, item_id, rule_type, action)
			VALUES (50, 'imagebbs', 'board', 50, 'allow', 'post'),
			       (50, 'images', 'upload', NULL, 'allow', 'post');
	`)
	require.NoError(t, err)

	session := sessions.NewSession(nil, "image-board")
	session.Values["UID"] = int32(50)
	q := db.NewForDriver(conn, "sqlite")
	queue := jobs.NewService(q)
	cd := common.NewCoreData(ctx, q, cfg, common.WithJobService(queue)).ForUser(50)
	cd.SetEvent(&eventbus.TaskEvent{})
	cd.SetSession(session)
	run := func() any {
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		require.NoError(t, form.WriteField("text", "Searchable image caption"))
		file, err := form.CreateFormFile("image", "image.png")
		require.NoError(t, err)
		require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 1, 1))))
		require.NoError(t, form.Close())
		req := httptest.NewRequest("POST", "/imagebbs/board/50", &body)
		req.Header.Set("Content-Type", form.FormDataContentType())
		req = req.WithContext(context.WithValue(ctx, consts.KeyCoreData, cd))
		req = mux.SetURLVars(req, map[string]string{"board": "50"})
		return uploadImageTask.Action(httptest.NewRecorder(), req)
	}

	return conn, q, queue, run
}

func TestAdminImagePostUpdateAndDeleteCommitDurableSearch(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite"))

	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (50, 'image admin');
		INSERT INTO imageboard (idimageboard, imageboard_idimageboard, title) VALUES (50, 0, 'Images');
		INSERT INTO imagepost (idimagepost, forumthread_id, users_idusers, imageboard_idimageboard, description, approved)
			VALUES (77, 1, 50, 50, 'Initial caption term', 1);
	`)
	require.NoError(t, err)

	q := db.NewForDriver(conn, "sqlite")
	queue := jobs.NewConfiguredService(q)
	require.NoError(t, searchworker.Index(ctx, q, searchworker.IndexEventData{
		Type: searchworker.TypeImage, ID: 77, Text: "Initial caption term",
	}))
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserID(50), common.WithJobService(queue))

	readImageWords := func() []string {
		t.Helper()
		rows, err := conn.QueryContext(ctx, `SELECT sw.word FROM searchwordlist sw
			JOIN imagepost_search ips ON ips.searchwordlist_idsearchwordlist = sw.idsearchwordlist
			WHERE ips.image_post_id = 77`)
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

	runDurableSearchPass := func(wantRev int64) {
		t.Helper()
		wCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			searchworker.RunDurableWorker(wCtx, queue, q, "image-search-worker")
		}()
		require.Eventually(t, func() bool {
			state, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
				TargetType: searchindex.TypeImage, TargetID: 77,
			})
			return err == nil && state.Revision == wantRev && state.IndexedRevision == wantRev
		}, 4*time.Second, 15*time.Millisecond)
		cancel()
		<-done
	}

	// 1. Trigger failure rolls back ModifyPostTask cleanly.
	_, err = conn.ExecContext(ctx, `CREATE TRIGGER reject_image_edit_search BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'image'
		BEGIN SELECT RAISE(ABORT, 'injected image edit failure'); END`)
	require.NoError(t, err)

	runModify := func(desc string) any {
		form := url.Values{"board": {"50"}, "desc": {desc}, "approved": {"1"}}
		req := httptest.NewRequest(http.MethodPost, "/admin/user/50/imagebbs/post/77", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"user": "50", "post": "77"})
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		return modifyPostTask.Action(httptest.NewRecorder(), req)
	}

	editErr, ok := runModify("UpdatedImageSearchToken").(error)
	require.True(t, ok)
	require.ErrorContains(t, editErr, "injected image edit failure")
	var curDesc string
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT description FROM imagepost WHERE idimagepost = 77`).Scan(&curDesc))
	require.Equal(t, "Initial caption term", curDesc)

	_, err = conn.ExecContext(ctx, `DROP TRIGGER reject_image_edit_search`)
	require.NoError(t, err)
	_ = runModify("UpdatedImageSearchToken")
	runDurableSearchPass(1)
	require.ElementsMatch(t, []string{"updatedimagesearchtoken"}, readImageWords())

	// 2. DeletePostTask commits durable search clearance and removes indexed words.
	req := httptest.NewRequest(http.MethodPost, "/admin/user/50/imagebbs/post/77/delete", nil)
	req = mux.SetURLVars(req, map[string]string{"user": "50", "post": "77"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	_ = deletePostTask.Action(httptest.NewRecorder(), req)
	runDurableSearchPass(2)
	require.Empty(t, readImageWords())
}
