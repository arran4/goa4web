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
	queries := &db.QuerierStub{
		GetPermissionsByUserIDReturns: []*db.GetPermissionsByUserIDRow{
			{Name: "moderator"},
		},
	}
	sm := &sessionManagerStub{}

	session := &sessions.Session{ID: "sessid", Values: map[interface{}]interface{}{"UID": int32(1)}}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
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
		WithEmailRegistry(reg),
		WithImageSigner(signer),
		WithLinkSigner(linkSigner),
		WithNavRegistry(navReg),
		WithQueries(queries),
		WithSessionManager(sm),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone", "user", "moderator"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if len(sm.insertCalls) != 1 || sm.insertCalls[0].sessionID != "sessid" || sm.insertCalls[0].userID != 1 {
		t.Fatalf("session insert not recorded, got %#v", sm.insertCalls)
	}
	if len(queries.GetPermissionsByUserIDCalls) != 1 || queries.GetPermissionsByUserIDCalls[0] != 1 {
		t.Fatalf("unexpected permission lookups: %#v", queries.GetPermissionsByUserIDCalls)
	}
}

func TestCoreDataMiddlewareAnonymous(t *testing.T) {
	navReg := nav.NewRegistry()

	cfg := config.NewRuntimeConfig()
	queries := &db.QuerierStub{
		GetPermissionsByUserIDReturns: []*db.GetPermissionsByUserIDRow{},
	}
	sm := &sessionManagerStub{}

	session := &sessions.Session{ID: "sessid"}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
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
		WithEmailRegistry(reg),
		WithImageSigner(signer),
		WithLinkSigner(linkSigner),
		WithNavRegistry(navReg),
		WithQueries(queries),
		WithSessionManager(sm),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if len(sm.deleteCalls) != 1 || sm.deleteCalls[0] != "sessid" {
		t.Fatalf("session delete not recorded, got %#v", sm.deleteCalls)
	}
}

type sessionCall struct {
	sessionID string
	userID    int32
}

type sessionManagerStub struct {
	insertCalls []sessionCall
	deleteCalls []string
}

func (s *sessionManagerStub) InsertSession(_ context.Context, sessionID string, userID int32) error {
	s.insertCalls = append(s.insertCalls, sessionCall{sessionID: sessionID, userID: userID})
	return nil
}

func (s *sessionManagerStub) DeleteSessionByID(_ context.Context, sessionID string) error {
	s.deleteCalls = append(s.deleteCalls, sessionID)
	return nil
}
