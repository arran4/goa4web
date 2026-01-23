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
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForgotPassword_VerifiedEmail(t *testing.T) {
	qs := testhelpers.NewQuerierStub()

	// Mock User Credentials
	qs.SystemGetLoginRow = &db.SystemGetLoginRow{
		Idusers: 101,
		Username: sql.NullString{String: "user_verified", Valid: true},
	}

	// Mock Verified Emails
	qs.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{
		{Email: "test@example.com"},
	}

	// Mock SystemGetUserByEmail
	qs.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{
		Idusers: 101,
	}

	// Mock Password Reset Creation
	qs.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
		return nil, sql.ErrNoRows
	}
	qs.CreatePasswordResetForUserFn = func(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
		return nil
	}
	qs.AdminInsertRequestQueueFn = func(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
		return db.FakeSQLResult{}, nil
	}

	form := url.Values{}
	form.Set("username", "user_verified")
	form.Set("password", "newPass123")

	req := httptest.NewRequest(http.MethodPost, "/login/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()

	// Use the exported type directly
	task := ForgotPasswordTask{}
	task.Action(rr, req) // Wait, Action returns 'any', we need to check return value or side effects.

	// Check if correct template handler is invoked.
	// Since Action returns the handler result, we can't easily check it unless we wrap it.
	// But we can check event data and DB calls.

	if len(qs.CreatePasswordResetForUserCalls) != 1 {
		t.Errorf("expected reset creation, got %d", len(qs.CreatePasswordResetForUserCalls))
	}

	if val, ok := evt.Data["UserHasNoVerifiedEmail"].(bool); ok && val {
		t.Errorf("expected UserHasNoVerifiedEmail to be false")
	}
}

func TestForgotPassword_NoVerifiedEmail(t *testing.T) {
	qs := testhelpers.NewQuerierStub()

	// Mock User Credentials
	qs.SystemGetLoginRow = &db.SystemGetLoginRow{
		Idusers: 102,
		Username: sql.NullString{String: "user_no_email", Valid: true},
	}

	// Mock No Verified Emails
	qs.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{}

	// Mock Password Reset Creation
	qs.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
		return nil, sql.ErrNoRows
	}
	qs.CreatePasswordResetForUserFn = func(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
		return nil
	}
	qs.AdminInsertRequestQueueFn = func(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
		return db.FakeSQLResult{}, nil
	}

	form := url.Values{}
	form.Set("username", "user_no_email")
	form.Set("password", "newPassRequest")

	req := httptest.NewRequest(http.MethodPost, "/login/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt))

	// Mock Template rendering by ensuring we don't panic and return valid handler
	// To verify which template was selected, we can check the return value of Action if we could inspect it.
	// But Action returns handlers.HandlerFunc result which is void.
	// However, we can inspect the response body if the template rendered?
	// But tests usually don't have templates loaded.
	// So ForgotPasswordRequestSentPageTmpl.Handler(...) will fail if we try to execute it without templates?
	// Or it returns a handler that writes to response.

	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()

	task := ForgotPasswordTask{}

	// We need to inject a template loader or mock templates?
	// The Action calls ForgotPasswordRequestSentPageTmpl.Handler.
	// That method does render using cd.Templates.
	// Since we didn't load templates in cd, it might error or panic?
	// common.NewCoreData doesn't load templates by default.
	// But we can check if Action logic reached the point of return.

	// Actually, CreatePasswordResetForUser being called is the main side effect.
	task.Action(rr, req)

	if len(qs.CreatePasswordResetForUserCalls) != 1 {
		t.Errorf("expected reset creation, got %d", len(qs.CreatePasswordResetForUserCalls))
	}

	if val, ok := evt.Data["UserHasNoVerifiedEmail"].(bool); !ok || !val {
		t.Errorf("expected UserHasNoVerifiedEmail to be true")
	}
}
