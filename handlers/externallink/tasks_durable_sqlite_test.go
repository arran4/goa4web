//go:build sqlite || sqlite3

package externallink

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/migrations"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func openDurableExternalLinkTestDB(t *testing.T) (*sql.DB, db.Querier, *jobs.Service) {
	t.Helper()
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite"))
	q := db.NewForDriver(conn, "sqlite")
	return conn, q, jobs.NewConfiguredService(q)
}

func TestDurablePrefetchAndReloadRouteThroughJobQueue(t *testing.T) {
	ctx := t.Context()
	conn, q, svc := openDurableExternalLinkTestDB(t)

	jobCount := func() int {
		var n int
		require.NoError(t, conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM durable_jobs`).Scan(&n))
		return n
	}

	// 1. Unauthenticated prefetch must be rejected with 401 and create zero durable jobs.
	anonCD := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithJobService(svc))
	anonCD.SetEvent(&eventbus.TaskEvent{})
	req := httptest.NewRequest(http.MethodPost, "/goto/prefetch?url="+url.QueryEscape("https://example.test/unauth"), nil)
	require.NoError(t, req.ParseForm())
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, anonCD))
	rec := httptest.NewRecorder()
	res := prefetchExternalLinkTask.Action(rec, req)
	h, ok := res.(http.Handler)
	require.True(t, ok, "expected http.Handler for 401 response, got %T", res)
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Zero(t, jobCount())

	// 2. Authenticated private/loopback URL must be rejected cleanly.
	authCD := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserID(12), common.WithJobService(svc))
	for _, bad := range []string{"http://127.0.0.1/secret", "http://localhost/internal"} {
		authCD.SetEvent(&eventbus.TaskEvent{})
		req := httptest.NewRequest(http.MethodPost, "/goto/prefetch?url="+url.QueryEscape(bad), nil)
		require.NoError(t, req.ParseForm())
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, authCD))
		out := prefetchExternalLinkTask.Action(httptest.NewRecorder(), req)
		require.Equal(t, handlers.TextByteWriter("invalid url"), out)
	}
	require.Zero(t, jobCount())

	// 3. Authenticated valid URL enqueues a durable external_link job and does NOT attach a BackgroundTasker.
	authCD.SetEvent(&eventbus.TaskEvent{})
	req = httptest.NewRequest(http.MethodPost, "/goto/prefetch?url="+url.QueryEscape("https://example.test/prefetch?utm_source=x&id=1"), nil)
	require.NoError(t, req.ParseForm())
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, authCD))
	out := prefetchExternalLinkTask.Action(httptest.NewRecorder(), req)
	require.Equal(t, handlers.TextByteWriter("ok"), out)
	_, isBg := authCD.Event().Task.(tasks.BackgroundTasker)
	require.False(t, isBg, "durable prefetch must not attach a volatile BackgroundTasker")

	listed, err := svc.AdminList(ctx, "", jobs.JobTypeExternalLink, 10, 0)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, "https://example.test/prefetch?id=1", listed[0].TargetKey)

	// 4. Signed ReloadExternalLinkTask enqueues a durable job without attaching BackgroundTasker, and rejects loopback.
	// signKey signs external-link goto URLs in the durable reload handler test.
	const signKey = "test-reload-sign-key"
	reloadCD := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserID(12), common.WithLinkSignKey(signKey), common.WithJobService(svc))
	reloadCD.SetEvent(&eventbus.TaskEvent{})
	validURL := "https://example.test/reload-page"
	sig := sign.Sign("link:"+validURL, signKey)
	reloadReq := httptest.NewRequest(http.MethodPost, "/goto?u="+url.QueryEscape(validURL)+"&sig="+sig, nil)
	require.NoError(t, reloadReq.ParseForm())
	reloadCD.LoadSelectionsFromRequest(reloadReq)
	reloadReq = reloadReq.WithContext(context.WithValue(reloadReq.Context(), consts.KeyCoreData, reloadCD))
	reloadRes := reloadExternalLinkTask.Action(httptest.NewRecorder(), reloadReq)
	_, isRedirect := reloadRes.(handlers.RedirectHandler)
	require.True(t, isRedirect, "expected RedirectHandler, got %T", reloadRes)
	_, isBg = reloadCD.Event().Task.(tasks.BackgroundTasker)
	require.False(t, isBg, "durable reload must not attach a volatile BackgroundTasker")

	listed, err = svc.AdminList(ctx, "", jobs.JobTypeExternalLink, 10, 0)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	require.Equal(t, validURL, listed[0].TargetKey)

	badCD := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserID(12), common.WithLinkSignKey(signKey), common.WithJobService(svc))
	badCD.SetEvent(&eventbus.TaskEvent{})
	loopURL := "http://127.0.0.1/internal"
	loopSig := sign.Sign("link:"+loopURL, signKey)
	badReq := httptest.NewRequest(http.MethodPost, "/goto?u="+url.QueryEscape(loopURL)+"&sig="+loopSig, nil)
	require.NoError(t, badReq.ParseForm())
	badCD.LoadSelectionsFromRequest(badReq)
	badReq = badReq.WithContext(context.WithValue(badReq.Context(), consts.KeyCoreData, badCD))
	badRes := reloadExternalLinkTask.Action(httptest.NewRecorder(), badReq)
	badErr, ok := badRes.(error)
	require.True(t, ok)
	require.ErrorIs(t, badErr, handlers.ErrForbidden)
	require.Equal(t, 2, jobCount())
}
