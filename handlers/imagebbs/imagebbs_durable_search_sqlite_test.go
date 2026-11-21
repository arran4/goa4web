//go:build sqlite || sqlite3

package imagebbs

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/png"
	"mime/multipart"
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
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestImageBoardUploadCommitsWithSearchIntent(t *testing.T) {
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
		CREATE TRIGGER reject_image_search BEFORE INSERT ON durable_jobs
			WHEN NEW.job_type = 'search_index' AND NEW.target_type = 'image'
			BEGIN SELECT RAISE(ABORT, 'injected image search failure'); END;
	`)
	require.NoError(t, err)

	cfg := config.NewRuntimeConfig()
	cfg.ImageMaxBytes = 1024 * 1024
	cfg.ImageUploadProvider = ""
	session := sessions.NewSession(nil, "image-board")
	session.Values["UID"] = int32(50)
	q := db.NewForDriver(conn, "sqlite")
	queue := jobs.NewService(q)
	cd := common.NewCoreData(ctx, q, cfg, common.WithJobService(queue)).ForUser(50)
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
