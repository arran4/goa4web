package auth

import (
	"context"
	"database/sql"
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

func TestLoginTask_Security_UsernameEnumeration(t *testing.T) {
	// Helper to extract error message from response body
	getErrorMsg := func(body string) string {
		if strings.Contains(body, "No such user") {
			return "No such user"
		}
		if strings.Contains(body, "Invalid password") {
			return "Invalid password"
		}
		if strings.Contains(body, "Invalid username or password") {
			return "Invalid username or password"
		}
		return "Unknown error: " + body
	}

	runLogin := func(username string, setupQuerier func(*db.QuerierStub)) string {
		q := testhelpers.NewQuerierStub()
		if setupQuerier != nil {
			setupQuerier(q)
		}

		form := url.Values{"username": {username}, "password": {"wrongpass"}}
		req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"

		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		return getErrorMsg(rr.Body.String())
	}

	t.Run("Error messages should be identical", func(t *testing.T) {
		// Case 1: Non-existent user
		msgNoUser := runLogin("nonexistent", func(q *db.QuerierStub) {
			q.SystemGetLoginErr = sql.ErrNoRows
		})

		// Case 2: Existing user, wrong password
		msgWrongPass := runLogin("existing", func(q *db.QuerierStub) {
			q.GetPasswordResetByUserErr = sql.ErrNoRows
			q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
				return &db.SystemGetLoginRow{
					Idusers:         1,
					Passwd:          sql.NullString{String: "somehash", Valid: true},
					PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
					Username:        sql.NullString{String: "existing", Valid: true},
				}, nil
			}
		})

		if msgNoUser != msgWrongPass {
			t.Errorf("Security vulnerability: Error messages differ.\nNo User: %q\nWrong Pass: %q", msgNoUser, msgWrongPass)
		}
	})
}
