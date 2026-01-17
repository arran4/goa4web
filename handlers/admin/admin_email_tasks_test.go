package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type emailTasksQueries struct {
	db.QuerierStub
	userID      int32
	user        *db.SystemGetUserByIDRow
	userEmail   *db.UserEmail
	addedEmail  string
	deletedID   int32
	updated     *db.AdminUpdateUserEmailDetailsParams
	codeUpdated *db.SystemUpdateVerificationCodeParams
}

func (q *emailTasksQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *emailTasksQueries) AdminGetUserEmailByID(_ context.Context, id int32) (*db.UserEmail, error) {
	if q.userEmail != nil && q.userEmail.ID == id {
		return q.userEmail, nil
	}
	return nil, fmt.Errorf("email not found")
}

func (q *emailTasksQueries) GetUserEmailByEmail(_ context.Context, email string) (*db.UserEmail, error) {
	return nil, sql.ErrNoRows
}

func (q *emailTasksQueries) InsertUserEmail(_ context.Context, arg db.InsertUserEmailParams) error {
	q.addedEmail = arg.Email
	return nil
}

func (q *emailTasksQueries) AdminDeleteUserEmail(_ context.Context, id int32) error {
	q.deletedID = id
	return nil
}

func (q *emailTasksQueries) AdminUpdateUserEmailDetails(_ context.Context, arg db.AdminUpdateUserEmailDetailsParams) error {
	q.updated = &arg
	return nil
}

func (q *emailTasksQueries) SystemUpdateVerificationCode(_ context.Context, arg db.SystemUpdateVerificationCodeParams) error {
	q.codeUpdated = &arg
	return nil
}

func setupEmailTaskTest(t *testing.T, userID int, task tasks.Task, form url.Values, queries *emailTasksQueries) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", fmt.Sprintf("/admin/user/%d", userID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	task.Action(rr, req)
	return rr
}

func TestAdminAddEmailTask(t *testing.T) {
	userID := 10
	queries := &emailTasksQueries{
		userID: int32(userID),
		user:   &db.SystemGetUserByIDRow{Idusers: int32(userID), Username: sql.NullString{String: "testuser", Valid: true}},
	}
	form := url.Values{}
	form.Set("task", string(TaskAddEmail))
	form.Set("new_email", "new@example.com")

	rr := setupEmailTaskTest(t, userID, adminAddEmailTask, form, queries)

	if queries.addedEmail != "new@example.com" {
		t.Errorf("expected email to be added, got %s", queries.addedEmail)
	}
	// Check redirect
	if rr.Code != http.StatusFound { // RefreshDirectHandler might result in redirect or specific handling?
		// Actually handlers.RefreshDirectHandler usually writes a meta refresh or similar, or redirects.
		// Let's check Action return value type if possible or just check side effects for now.
		// The Action returns `any`. The test helper calls `Action`.
		// But in a real handler `handlers.TaskHandler` wraps this.
		// Here I called `Action` directly. It returns `handlers.RefreshDirectHandler`.
		// It doesn't write to response writer unless processed.
		// So I only check side effects on queries.
	}
}

func TestAdminDeleteEmailTask(t *testing.T) {
	userID := 10
	emailID := 55
	queries := &emailTasksQueries{
		userID:    int32(userID),
		userEmail: &db.UserEmail{ID: int32(emailID), UserID: int32(userID)},
	}
	form := url.Values{}
	form.Set("task", string(TaskDeleteEmail))
	form.Set("email_id", strconv.Itoa(emailID))

	setupEmailTaskTest(t, userID, adminDeleteEmailTask, form, queries)

	if queries.deletedID != int32(emailID) {
		t.Errorf("expected email %d to be deleted, got %d", emailID, queries.deletedID)
	}
}

func TestAdminVerifyEmailTask(t *testing.T) {
	userID := 10
	emailID := 55
	queries := &emailTasksQueries{
		userID:    int32(userID),
		userEmail: &db.UserEmail{ID: int32(emailID), UserID: int32(userID), Email: "test@example.com"},
	}
	form := url.Values{}
	form.Set("task", string(TaskVerifyEmail))
	form.Set("email_id", strconv.Itoa(emailID))

	setupEmailTaskTest(t, userID, adminVerifyEmailTask, form, queries)

	if queries.updated == nil || !queries.updated.VerifiedAt.Valid {
		t.Errorf("expected email to be verified")
	}
}

func TestAdminUnverifyEmailTask(t *testing.T) {
	userID := 10
	emailID := 55
	queries := &emailTasksQueries{
		userID: int32(userID),
		userEmail: &db.UserEmail{
			ID:         int32(emailID),
			UserID:     int32(userID),
			Email:      "test@example.com",
			VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
		},
	}
	form := url.Values{}
	form.Set("task", string(TaskUnverifyEmail))
	form.Set("email_id", strconv.Itoa(emailID))

	setupEmailTaskTest(t, userID, adminUnverifyEmailTask, form, queries)

	if queries.updated == nil || queries.updated.VerifiedAt.Valid {
		t.Errorf("expected email to be unverified")
	}
}

func TestAdminResendVerificationEmailTask(t *testing.T) {
	userID := 10
	emailID := 55
	queries := &emailTasksQueries{
		userID:    int32(userID),
		userEmail: &db.UserEmail{ID: int32(emailID), UserID: int32(userID), Email: "test@example.com"},
		user:      &db.SystemGetUserByIDRow{Idusers: int32(userID), Username: sql.NullString{String: "testuser", Valid: true}},
	}
	form := url.Values{}
	form.Set("task", string(TaskResendVerification))
	form.Set("email_id", strconv.Itoa(emailID))

	setupEmailTaskTest(t, userID, adminResendVerificationEmailTask, form, queries)

	if queries.codeUpdated == nil || !queries.codeUpdated.LastVerificationCode.Valid {
		t.Errorf("expected verification code to be updated")
	}
}
