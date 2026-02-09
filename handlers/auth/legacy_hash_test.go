package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/sessions"
)

func TestMD5LoginRejected(t *testing.T) {
	// Setup CoreData with a QuerierStub that returns an MD5 hashed password
	q := testhelpers.NewQuerierStub()
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		// md5("secret") = 5ebe2294ecd0e0f08eab7690d2a6ee69
		return &db.SystemGetLoginRow{
			Idusers:         1,
			Passwd:          sql.NullString{String: "5ebe2294ecd0e0f08eab7690d2a6ee69", Valid: true},
			PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
			Username:        sql.NullString{String: "olduser", Valid: true},
		}, nil
	}
	q.GetLoginRoleForUserReturns = 1
	// The login task will try to check for password reset if login fails.
	// We want to simulate no reset pending, so it returns "Invalid password".
	q.GetPasswordResetByUserErr = sql.ErrNoRows

	// Init sessions
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	form := url.Values{"username": {"olduser"}, "password": {"secret"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	// MD5 login should fail
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK (login page re-rendered), got %d. Body: %s", rr.Code, rr.Body.String())
	}
	// Check if it's a failure
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid password") {
		t.Fatalf("Expected login failure message 'Invalid password', got: %s", body)
	}
	if strings.Contains(cd.AutoRefresh, "url=/") {
		t.Fatalf("Did not expect success redirect")
	}
}
