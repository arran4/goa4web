package user

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

type timeCloseMatcher struct {
	target time.Time
	margin time.Duration
}

func (m timeCloseMatcher) Match(v driver.Value) bool {
	t, ok := v.(time.Time)
	if !ok {
		return false
	}
	if t.After(m.target) {
		return t.Sub(m.target) <= m.margin
	}
	return m.target.Sub(t) <= m.margin
}

func TestAddEmailTaskUsesConfigExpiry(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
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
	codeMatcher := sqlmock.AnyArg()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority\nFROM user_emails\nWHERE email = ?")).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_emails (user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority)\nVALUES (?, ?, ?, ?, ?, ?)\n")).
		WithArgs(int32(1), "new@example.com", sqlmock.AnyArg(), codeMatcher, timeCloseMatcher{target: expectedExpire, margin: 2 * time.Second}, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at\nFROM users u\nLEFT JOIN user_emails ue ON ue.id = (\n        SELECT id FROM user_emails ue2\n        WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL\n        ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1\n)\nWHERE u.idusers = ?\n")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, sql.NullString{String: "primary@example.com", Valid: true}, sql.NullString{String: "tester", Valid: true}, sql.NullTime{}))

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
