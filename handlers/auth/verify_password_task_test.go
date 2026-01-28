package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestVerifyPasswordAction_Success(t *testing.T) {
	pwHash, alg, _ := HashPassword("pw")
	reset := &db.PendingPassword{
		ID:               1,
		UserID:           1,
		Passwd:           pwHash,
		PasswdAlgorithm:  alg,
		VerificationCode: "code",
		CreatedAt:        time.Now(),
	}
	q := testhelpers.NewQuerierStub()
	q.GetPasswordResetByCodeReturns = reset
	q.GetPendingPasswordByCodeReturns = reset
	q.GetLoginRoleForUserReturns = 1
	q.SystemMarkPasswordResetVerifiedErr = nil
	q.InsertPasswordErr = nil

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	form := url.Values{"id": {"1"}, "code": {"code"}, "password": {"pw"}}
	req := httptest.NewRequest(http.MethodPost, "/login/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(verifyPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(q.GetPasswordResetByCodeCalls) != 1 {
		t.Fatalf("expected GetPasswordResetByCode to be called once, got %d", len(q.GetPasswordResetByCodeCalls))
	}
	if len(q.SystemMarkPasswordResetVerifiedCalls) != 1 {
		t.Fatalf("expected SystemMarkPasswordResetVerified to be called once, got %d", len(q.SystemMarkPasswordResetVerifiedCalls))
	}
	if len(q.InsertPasswordCalls) != 1 {
		t.Fatalf("expected InsertPassword to be called once, got %d", len(q.InsertPasswordCalls))
	}
}

func TestVerifyPasswordAction_InvalidPassword(t *testing.T) {
	pwHash, alg, _ := HashPassword("pw")
	reset := &db.PendingPassword{
		ID:               1,
		UserID:           1,
		Passwd:           pwHash,
		PasswdAlgorithm:  alg,
		VerificationCode: "code",
		CreatedAt:        time.Now(),
	}
	q := testhelpers.NewQuerierStub()
	q.GetPasswordResetByCodeReturns = reset
	q.GetPendingPasswordByCodeReturns = reset
	q.GetLoginRoleForUserReturns = 1
	q.SystemMarkPasswordResetVerifiedErr = nil
	q.InsertPasswordErr = nil

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	form := url.Values{"id": {"1"}, "code": {"code"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(verifyPasswordTask)(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(q.SystemMarkPasswordResetVerifiedCalls) != 0 {
		t.Fatalf("expected SystemMarkPasswordResetVerified not to be called, got %d", len(q.SystemMarkPasswordResetVerifiedCalls))
	}
	if len(q.InsertPasswordCalls) != 0 {
		t.Fatalf("expected InsertPassword not to be called, got %d", len(q.InsertPasswordCalls))
	}
}
