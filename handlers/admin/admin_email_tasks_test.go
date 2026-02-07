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
	"github.com/arran4/goa4web/internal/testhelpers"
)

func setupEmailTaskTest(t *testing.T, userID int, task tasks.Task, form url.Values, queries *db.QuerierStub) *httptest.ResponseRecorder {
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
	t.Run("Happy Path", func(t *testing.T) {
		userID := 10
		qs := testhelpers.NewQuerierStub()
		qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{Idusers: int32(userID), Username: sql.NullString{String: "testuser", Valid: true}}

		form := url.Values{}
		form.Set("task", string(TaskAddEmail))
		form.Set("new_email", "new@example.com")

		rr := setupEmailTaskTest(t, userID, adminAddEmailTask, form, qs)

		if len(qs.InsertUserEmailCalls) != 1 {
			t.Errorf("expected email to be added")
		} else if qs.InsertUserEmailCalls[0].Email != "new@example.com" {
			t.Errorf("expected email to be added, got %s", qs.InsertUserEmailCalls[0].Email)
		}

		if rr.Code != http.StatusFound {
			// ...
		}
	})
}

func TestAdminDeleteEmailTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		userID := 10
		emailID := 55
		qs := testhelpers.NewQuerierStub()
		qs.AdminGetUserEmailByIDReturns = &db.UserEmail{ID: int32(emailID), UserID: int32(userID)}

		form := url.Values{}
		form.Set("task", string(TaskDeleteEmail))
		form.Set("email_id", strconv.Itoa(emailID))

		setupEmailTaskTest(t, userID, adminDeleteEmailTask, form, qs)

		if len(qs.AdminDeleteUserEmailCalls) != 1 {
			t.Errorf("expected email to be deleted")
		} else if qs.AdminDeleteUserEmailCalls[0] != int32(emailID) {
			t.Errorf("expected email %d to be deleted, got %d", emailID, qs.AdminDeleteUserEmailCalls[0])
		}
	})
}

func TestAdminVerifyEmailTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		userID := 10
		emailID := 55
		qs := testhelpers.NewQuerierStub()
		qs.AdminGetUserEmailByIDReturns = &db.UserEmail{ID: int32(emailID), UserID: int32(userID), Email: "test@example.com"}

		form := url.Values{}
		form.Set("task", string(TaskVerifyEmail))
		form.Set("email_id", strconv.Itoa(emailID))

		setupEmailTaskTest(t, userID, adminVerifyEmailTask, form, qs)

		if len(qs.AdminUpdateUserEmailDetailsCalls) != 1 {
			t.Errorf("expected email to be verified")
		} else if !qs.AdminUpdateUserEmailDetailsCalls[0].VerifiedAt.Valid {
			t.Errorf("expected verified_at to be set")
		}
	})
}

func TestAdminUnverifyEmailTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		userID := 10
		emailID := 55
		qs := testhelpers.NewQuerierStub()
		qs.AdminGetUserEmailByIDReturns = &db.UserEmail{
			ID:         int32(emailID),
			UserID:     int32(userID),
			Email:      "test@example.com",
			VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		form := url.Values{}
		form.Set("task", string(TaskUnverifyEmail))
		form.Set("email_id", strconv.Itoa(emailID))

		setupEmailTaskTest(t, userID, adminUnverifyEmailTask, form, qs)

		if len(qs.AdminUpdateUserEmailDetailsCalls) != 1 {
			t.Errorf("expected email to be unverified")
		} else if qs.AdminUpdateUserEmailDetailsCalls[0].VerifiedAt.Valid {
			t.Errorf("expected verified_at to be unset")
		}
	})
}

func TestAdminResendVerificationEmailTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		userID := 10
		emailID := 55
		qs := testhelpers.NewQuerierStub()
		qs.AdminGetUserEmailByIDReturns = &db.UserEmail{ID: int32(emailID), UserID: int32(userID), Email: "test@example.com"}
		qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{Idusers: int32(userID), Username: sql.NullString{String: "testuser", Valid: true}}

		form := url.Values{}
		form.Set("task", string(TaskResendVerification))
		form.Set("email_id", strconv.Itoa(emailID))

		setupEmailTaskTest(t, userID, adminResendVerificationEmailTask, form, qs)

		if len(qs.SystemUpdateVerificationCodeCalls) != 1 {
			t.Errorf("expected verification code to be updated")
		} else if !qs.SystemUpdateVerificationCodeCalls[0].LastVerificationCode.Valid {
			t.Errorf("expected verification code to be set")
		}
	})
}
