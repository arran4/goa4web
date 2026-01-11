package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

type addEmailQueries struct {
	db.Querier
	userID   int32
	user     *db.SystemGetUserByIDRow
	inserted []db.InsertUserEmailParams
}

func (q *addEmailQueries) GetUserEmailByEmail(_ context.Context, email string) (*db.UserEmail, error) {
	if email != "new@example.com" {
		return nil, fmt.Errorf("unexpected email lookup: %s", email)
	}
	return nil, sql.ErrNoRows
}

func (q *addEmailQueries) InsertUserEmail(_ context.Context, arg db.InsertUserEmailParams) error {
	q.inserted = append(q.inserted, arg)
	return nil
}

func (q *addEmailQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func TestAddEmailTaskUsesConfigExpiry(t *testing.T) {
	queries := &addEmailQueries{
		userID: 1,
		user: &db.SystemGetUserByIDRow{
			Idusers:                1,
			Email:                  sql.NullString{String: "primary@example.com", Valid: true},
			Username:               sql.NullString{String: "tester", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cfg := config.NewRuntimeConfig()
	cfg.EmailVerificationExpiryHours = 72
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, cfg, common.WithSession(sess), common.WithEvent(evt))
	cd.UserID = 1
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	expectedExpire := time.Now().Add(time.Duration(cfg.EmailVerificationExpiryHours) * time.Hour)

	form := url.Values{"new_email": {"new@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/usr/email", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(addEmailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}

	expiresAt, ok := evt.Data["ExpiresAt"].(time.Time)
	if !ok {
		t.Fatalf("expected expiry time in event data, got %#v", evt.Data["ExpiresAt"])
	}
	margin := 2 * time.Second
	if expiresAt.Before(expectedExpire.Add(-margin)) || expiresAt.After(expectedExpire.Add(margin)) {
		t.Fatalf("expiry %v not within margin of expected %v", expiresAt, expectedExpire)
	}

	if len(queries.inserted) != 1 {
		t.Fatalf("expected insert, got %d", len(queries.inserted))
	}
	inserted := queries.inserted[0]
	if inserted.UserID != 1 || inserted.Email != "new@example.com" {
		t.Fatalf("unexpected insert args: %#v", inserted)
	}
	if inserted.VerificationExpiresAt.Time.Before(expectedExpire.Add(-margin)) || inserted.VerificationExpiresAt.Time.After(expectedExpire.Add(margin)) {
		t.Fatalf("insert expiry %v not within margin of expected %v", inserted.VerificationExpiresAt.Time, expectedExpire)
	}
}
