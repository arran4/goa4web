package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// Mocking Admin Task needs access to admin package internals or just testing the flow logic via handlers?
// admin package is separate. I cannot access unexported types.
// But I can run `admin.UserCreateResetLinkTask` via `handlers.TaskHandler` if I can access it.
// Wait, `admin.UserCreateResetLinkTask` is NOT exported. I cannot access it from `auth` package.
// I should put the E2E test in a separate package `handlers/e2e_test` or similar, or just test `auth` parts here.
// The user asked for "a decent end to end test".
// If I put it in `handlers/auth/reset_password_test.go`, I cannot test Admin generation directly if the task is private.
// However, the *logic* is in `CoreData`. I can test `CoreData` logic.
// But the prompt implies testing the web flow.
// I can make `UserCreateResetLinkTask` exported in `handlers/admin`.
// Or I can test it in `handlers/admin/user_create_reset_link_task_test.go`.

// Let's split the test.
// `handlers/auth/reset_password_test.go` tests the User side (Consumption of link).
// `handlers/admin/user_reset_link_test.go` tests the Admin side (Generation).

// Wait, I can't easily pass data between two test files in different packages.
// But I can manually construct a signed link in `auth` test using `cd.SignPasswordResetLink`.
// That simulates the Admin action.

func TestResetPasswordFlow_AdminLink(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	// Setup CoreData
	cfg := config.NewRuntimeConfig()
	cfg.HTTPHostname = "http://example.com"
	cd := common.NewCoreData(context.Background(), q, cfg, common.WithLinkSignKey("secret"))
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	// 1. Admin generates link (Simulated via CoreData)
	userID := int32(101)
	expiry := 24 * time.Hour

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(userID, sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO pending_passwords")).
		WithArgs(userID, sql.NullString{}, sql.NullString{}, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	genCode, err := cd.CreatePasswordResetForUser(userID, "", "")
	if err != nil {
		t.Fatalf("CreatePasswordResetForUser: %v", err)
	}

	// Sign Link
	link := cd.SignPasswordResetLink(genCode, expiry)
	// Link format: /login/reset?code=...&sig=...&ts=...
	u, _ := url.Parse(link)
	sig := u.Query().Get("sig")
	ts := u.Query().Get("ts")

	// 2. User GETs /login/reset
	req := httptest.NewRequest(http.MethodGet, link, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	ResetPasswordPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ResetPasswordPage status %d body: %s", rr.Code, rr.Body.String())
	}
	// Verify template data contains code, sig, ts
	// We can't easily check template data as it renders.
	// But if status is OK, signature verification passed.

	// 3. User POSTs /login/verify
	newPass := "newPassword123"
	form := url.Values{
		"code": {genCode},
		"sig": {sig},
		"ts": {ts},
		"password": {newPass},
	}
	reqPost := httptest.NewRequest(http.MethodPost, "/login/verify", strings.NewReader(form.Encode()))
	reqPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPost = reqPost.WithContext(ctx)
	rrPost := httptest.NewRecorder()

	// Mock expectations for Verify
	// 1. VerifyPasswordResetLink (Pure logic, no DB)
	// 2. VerifyPasswordReset (DB)
	//    a. GetPasswordResetByCode (With ZeroTime for expiryLookback because sig is present)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).
		WithArgs(genCode, sqlmock.AnyArg()). // CreatedAt arg is ZeroTime (0001-01-01...)
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
			AddRow(1, userID, sql.NullString{}, sql.NullString{}, genCode, time.Now(), nil))

	//    b. GetLoginRoleForUser
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles")).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"col"}).AddRow(1))

	//    c. Mark Verified
	mock.ExpectExec(regexp.QuoteMeta("UPDATE pending_passwords")).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	//    d. Insert Password (Hash logic)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO passwords")).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()). // Hashed password
		WillReturnResult(sqlmock.NewResult(1, 1))

	handlers.TaskHandler(verifyPasswordTask)(rrPost, reqPost)

	if rrPost.Code != http.StatusOK && rrPost.Code != http.StatusSeeOther {
		// TaskHandler returns RefreshDirectHandler which renders TaskDoneAutoRefreshPage (200 OK)
		// Or redirects?
		t.Fatalf("VerifyPasswordTask status %d", rrPost.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
