package auth

import (
	"context"
	"database/sql"
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

func TestVerifyPasswordTask_Action(t *testing.T) {
	t.Run("Happy Path - Success", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()

		pwHash, alg, _ := HashPassword("pw")
		q.GetPasswordResetByCodeReturns = &db.PendingPassword{
			ID:               1,
			UserID:           1,
			Passwd:           pwHash,
			PasswdAlgorithm:  alg,
			VerificationCode: "code",
			CreatedAt:        time.Now(),
			VerifiedAt:       sql.NullTime{},
		}
		q.GetLoginRoleForUserReturns = 1

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
		if len(q.SystemMarkPasswordResetVerifiedCalls) != 1 {
			t.Fatalf("expected reset verified")
		}
		if len(q.InsertPasswordCalls) != 1 {
			t.Fatalf("expected password insert")
		}
	})

	t.Run("Unhappy Path - Invalid Password", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()

		pwHash, alg, _ := HashPassword("pw")
		q.GetPasswordResetByCodeReturns = &db.PendingPassword{
			ID:               1,
			UserID:           1,
			Passwd:           pwHash,
			PasswdAlgorithm:  alg,
			VerificationCode: "code",
			CreatedAt:        time.Now(),
			VerifiedAt:       sql.NullTime{},
		}
		q.GetLoginRoleForUserReturns = 1

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
			t.Fatalf("unexpected reset verified")
		}
	})
}
