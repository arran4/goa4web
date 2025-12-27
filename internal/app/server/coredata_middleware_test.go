package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	imagesign "github.com/arran4/goa4web/internal/images"
	linksign "github.com/arran4/goa4web/internal/linksign"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"
)

func TestCoreDataMiddlewareUserRoles(t *testing.T) {
	navReg := nav.NewRegistry()

	cfg := config.NewRuntimeConfig()
	q := querierStub{
		roles: []*db.GetPermissionsByUserIDRow{
			{Name: "moderator"},
		},
	}

	session := &sessions.Session{ID: "sessid", Values: map[interface{}]interface{}{"UID": int32(1)}}
	req := httptest.NewRequest("GET", "/", nil)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	reg := email.NewRegistry()
	signer := imagesign.NewSigner(cfg, "k")
	linkSigner := linksign.NewSigner(cfg, "k")
	srv := New(
		WithConfig(cfg),
		WithQuerier(q),
		WithSessionManager(sessionManagerStub{}),
		WithEmailRegistry(reg),
		WithImageSigner(signer),
		WithLinkSigner(linkSigner),
		WithNavRegistry(navReg),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone", "user", "moderator"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}
}

func TestCoreDataMiddlewareAnonymous(t *testing.T) {
	navReg := nav.NewRegistry()

	cfg := config.NewRuntimeConfig()
	q := querierStub{}

	session := &sessions.Session{ID: "sessid"}
	req := httptest.NewRequest("GET", "/", nil)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	reg := email.NewRegistry()
	signer := imagesign.NewSigner(cfg, "k")
	linkSigner := linksign.NewSigner(cfg, "k")
	srv := New(
		WithConfig(cfg),
		WithQuerier(q),
		WithSessionManager(sessionManagerStub{}),
		WithEmailRegistry(reg),
		WithImageSigner(signer),
		WithLinkSigner(linkSigner),
		WithNavRegistry(navReg),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}
}

type querierStub struct {
	db.Querier
	roles []*db.GetPermissionsByUserIDRow
}

func (q querierStub) GetPermissionsByUserID(ctx context.Context, userID int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return q.roles, nil
}

func (querierStub) GetUnreadNotificationCountForLister(context.Context, int32) (int64, error) {
	return 0, nil
}

type sessionManagerStub struct{}

func (sessionManagerStub) InsertSession(context.Context, string, int32) error { return nil }
func (sessionManagerStub) DeleteSessionByID(context.Context, string) error    { return nil }
