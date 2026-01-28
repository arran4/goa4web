package server

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"
)

type sessionManagerStub struct {
	inserted []sessionInsert
	deleted  []string
}

type sessionInsert struct {
	sessionID string
	userID    int32
}

func (s *sessionManagerStub) InsertSession(_ context.Context, sessionID string, userID int32) error {
	s.inserted = append(s.inserted, sessionInsert{sessionID: sessionID, userID: userID})
	return nil
}

func (s *sessionManagerStub) DeleteSessionByID(_ context.Context, sessionID string) error {
	s.deleted = append(s.deleted, sessionID)
	return nil
}

func TestCoreDataMiddlewareUserRoles(t *testing.T) {
	navReg := nav.NewRegistry()
	cfg := config.NewRuntimeConfig()
	sm := &sessionManagerStub{}
	queries := testhelpers.NewQuerierStub(testhelpers.WithPermissions([]*db.GetPermissionsByUserIDRow{
		{
			IduserRoles:  1,
			UsersIdusers: 1,
			RoleID:       2,
			Name:         "moderator",
			IsAdmin:      false,
		},
	}))

	session := &sessions.Session{ID: "sessid", Values: map[interface{}]interface{}{"UID": int32(1)}}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	reg := email.NewRegistry()
	signer := "k"
	linkSigner := "k"
	srv := New(
		WithQuerier(queries),
		WithConfig(cfg),
		WithEmailRegistry(reg),
		WithImageSignKey(signer),
		WithLinkSignKey(linkSigner),
		WithNavRegistry(navReg),
		WithSessionManager(sm),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone", "user", "moderator"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if len(sm.inserted) != 1 {
		t.Fatalf("expected one session insert, got %d", len(sm.inserted))
	}
	if sm.inserted[0] != (sessionInsert{sessionID: "sessid", userID: 1}) {
		t.Fatalf("unexpected session insert: %+v", sm.inserted[0])
	}
}

func TestCoreDataMiddlewareAnonymous(t *testing.T) {
	navReg := nav.NewRegistry()
	cfg := config.NewRuntimeConfig()
	sm := &sessionManagerStub{}
	queries := testhelpers.NewQuerierStub()
	queries.SystemCheckGrantErr = sql.ErrNoRows

	session := &sessions.Session{ID: "sessid"}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	reg := email.NewRegistry()
	signer := "k"
	linkSigner := "k"
	srv := New(
		WithQuerier(queries),
		WithConfig(cfg),
		WithEmailRegistry(reg),
		WithImageSignKey(signer),
		WithLinkSignKey(linkSigner),
		WithNavRegistry(navReg),
		WithSessionManager(sm),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if len(sm.deleted) != 1 {
		t.Fatalf("expected one session delete, got %d", len(sm.deleted))
	}
	if sm.deleted[0] != "sessid" {
		t.Fatalf("unexpected session delete: %s", sm.deleted[0])
	}
}
