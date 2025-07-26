package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/gorilla/sessions"
)

func signBackURL(key, u string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(key))
	io.WriteString(mac, fmt.Sprintf("back:%s:%d", u, ts))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestLoginAction_NoSuchUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO login_attempts (username, ip_address) VALUES (?, ?)")).WithArgs("bob", "1.2.3.4").WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{"username": {"bob"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "No such user") {
		t.Fatalf("body=%q", body)
	}
}

func TestLoginAction_InvalidPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "e", "7c4f29407893c334a6cb7a87bf045c0d", "md5", "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnRows(rows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(int32(1), sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO login_attempts (username, ip_address) VALUES (?, ?)")).WithArgs("bob", "1.2.3.4").WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{"username": {"bob"}, "password": {"wrong"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid password") {
		t.Fatalf("body=%q", body)
	}
}

func TestLoginPageHiddenFields(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	q := dbpkg.New(db)

	req := httptest.NewRequest(http.MethodGet, "/login?code=abc&back=%2Ffoo&method=POST&data=x", nil)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	loginTask.Page(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "name=\"code\" value=\"abc\"") {
		t.Fatalf("missing code field: %q", body)
	}
	if !strings.Contains(body, "name=\"back\" value=\"/foo\"") {
		t.Fatalf("missing back field: %q", body)
	}
	if !strings.Contains(body, "name=\"method\" value=\"POST\"") {
		t.Fatalf("missing method field: %q", body)
	}
	if !strings.Contains(body, "name=\"data\" value=\"x\"") {
		t.Fatalf("missing data field: %q", body)
	}
	if strings.Contains(body, "back_sig") || strings.Contains(body, "back_ts") {
		t.Fatalf("unexpected signature fields: %q", body)
	}
}

func TestLoginAction_PendingResetPrompt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	pwHash, alg, _ := HashPassword("newpw")
	userRows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "e", "oldhash", "md5", "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnRows(userRows)
	resetRows := sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
		AddRow(2, 1, pwHash, alg, "code", time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(int32(1), sqlmock.AnyArg()).WillReturnRows(resetRows)

	form := url.Values{"username": {"bob"}, "password": {"newpw"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"

	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "name=\"id\" value=\"2\"") {
		t.Fatalf("missing id field: %q", body)
	}
}

func TestSanitizeBackURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	cfg.HTTPHostname = ""
	cd := common.NewCoreData(req.Context(), dbpkg.New(nil), cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if got := cd.SanitizeBackURL(req, "/foo"); got != "/foo" {
		t.Fatalf("relative got %q", got)
	}
	if got := cd.SanitizeBackURL(req, "https://example.com/bar?x=1"); got != "/bar?x=1" {
		t.Fatalf("host match got %q", got)
	}
	if got := cd.SanitizeBackURL(req, "https://evil.com/"); got != "" {
		t.Fatalf("evil got %q", got)
	}

	cfg.HTTPHostname = "https://example.com"
	cd = common.NewCoreData(req.Context(), dbpkg.New(nil), cfg)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	if got := cd.SanitizeBackURL(req, "https://example.com/baz"); got != "/baz" {
		t.Fatalf("cfg host got %q", got)
	}
}

func TestSanitizeBackURLSigned(t *testing.T) {
	raw := "https://evil.com/x"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	signer := imagesign.NewSigner(cfg, "k")
	cd := common.NewCoreData(req.Context(), dbpkg.New(nil), config.NewRuntimeConfig(), common.WithImageSigner(signer))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	ts := time.Now().Add(time.Hour).Unix()
	sig := signBackURL("k", raw, ts)
	q := req.URL.Query()
	q.Set("back_ts", fmt.Sprint(ts))
	q.Set("back_sig", sig)
	req.URL.RawQuery = q.Encode()
	if got := cd.SanitizeBackURL(req, raw); got != raw {
		t.Fatalf("signed got %q", got)
	}
}

func TestLoginPageInvalidBackURL(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	q := dbpkg.New(db)

	req := httptest.NewRequest(http.MethodGet, "/login?back=https://evil.com/x", nil)
	req.Host = "example.com"
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	loginTask.Page(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "name=\"back\"") {
		t.Fatalf("back field present: %q", body)
	}
}

func TestLoginPageSignedBackURL(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	q := dbpkg.New(db)

	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	raw := "https://evil.com/x"
	ts := time.Now().Add(time.Hour).Unix()
	sig := signBackURL("k", raw, ts)
	req := httptest.NewRequest(http.MethodGet, "/login?back="+url.QueryEscape(raw)+"&back_ts="+fmt.Sprint(ts)+"&back_sig="+sig, nil)
	req.Host = "example.com"
	signer := imagesign.NewSigner(cfg, "k")
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithImageSigner(signer))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	loginTask.Page(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "name=\"back\" value=\""+raw+"\"") {
		t.Fatalf("missing back field: %q", body)
	}
	if !strings.Contains(body, "name=\"back_sig\" value=") {
		t.Fatalf("missing back_sig field: %q", body)
	}
	if !strings.Contains(body, "name=\"back_ts\" value=") {
		t.Fatalf("missing back_ts field: %q", body)
	}
}

func TestLoginAction_ExternalBackURLIgnored(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	q := dbpkg.New(db)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	pwHash, alg, _ := HashPassword("pw")
	userRows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "e", pwHash, alg, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnRows(userRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {"https://evil.com"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	req.Host = "example.com"

	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "url=/") {
		t.Fatalf("missing refresh to root: %q", body)
	}
}

func TestLoginAction_SignedExternalBackURL(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	q := dbpkg.New(db)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM login_attempts")).
		WithArgs("bob", "1.2.3.4", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	pwHash, alg, _ := HashPassword("pw")
	userRows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "e", pwHash, alg, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnRows(userRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM login_attempts")).
		WithArgs("bob", "1.2.3.4", sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	raw := "https://example.org/ok"
	ts := time.Now().Add(time.Hour).Unix()
	sig := signBackURL("k", raw, ts)
	signer := imagesign.NewSigner(cfg, "k")
	form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {raw}, "back_ts": {fmt.Sprint(ts)}, "back_sig": {sig}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	req.Host = "example.com"

	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithImageSigner(signer))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != raw {
		t.Fatalf("location=%q", loc)
	}
}

func TestLoginAction_Throttle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM login_attempts")).
		WithArgs("bob", "1.2.3.4", sqlmock.AnyArg()).WillReturnRows(rows)

	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	cfg.LoginAttemptThreshold = 3
	cfg.LoginAttemptWindow = 15

	form := url.Values{"username": {"bob"}, "password": {"pw"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(loginTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Too many failed attempts") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}
