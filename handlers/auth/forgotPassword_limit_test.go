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

func TestForgotPasswordTask_Action(t *testing.T) {
	t.Run("Unhappy Path - Rate Limit Reached", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		now := time.Now()

		q.SystemGetLoginRow = &db.SystemGetLoginRow{
			Idusers: 1,
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		}

		q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{{
			UserID: 1,
			Email:  "a@test.com",
		}}

		q.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{
			Idusers: 1,
			Email:   "a@test.com",
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		}

		q.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
			return &db.PendingPassword{
				ID:               1,
				UserID:           1,
				Passwd:           "hash",
				PasswdAlgorithm:  "alg",
				VerificationCode: "code",
				CreatedAt:        now,
				VerifiedAt:       sql.NullTime{},
			}, nil
		}

		form := url.Values{"username": {"u"}, "password": {"pw"}}
		req := httptest.NewRequest(http.MethodPost, "/forgot", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handlers.TaskHandler(forgotPasswordTask)(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.CreatePasswordResetForUserCalls) != 0 {
			t.Fatalf("expected no reset inserts when rate limited, got %d", len(q.CreatePasswordResetForUserCalls))
		}
		if len(q.SystemDeletePasswordResetCalls) != 0 {
			t.Fatalf("unexpected delete of reset id")
		}
	})

	t.Run("Happy Path - Replace Old", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()

		q.SystemGetLoginRow = &db.SystemGetLoginRow{
			Idusers: 1,
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		}

		q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{{
			UserID: 1,
			Email:  "a@test.com",
		}}

		q.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{
			Idusers: 1,
			Email:   "a@test.com",
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		}

		q.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
			return &db.PendingPassword{
				ID:               1,
				UserID:           1,
				Passwd:           "hash",
				PasswdAlgorithm:  "alg",
				VerificationCode: "code",
				CreatedAt:        time.Now().Add(-25 * time.Hour),
				VerifiedAt:       sql.NullTime{},
			}, nil
		}

		q.SystemDeletePasswordResetFn = func(ctx context.Context, id int32) error {
			return nil
		}
		q.CreatePasswordResetForUserFn = func(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
			return nil
		}

		form := url.Values{"username": {"u"}, "password": {"pw"}}
		req := httptest.NewRequest(http.MethodPost, "/forgot", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handlers.TaskHandler(forgotPasswordTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemDeletePasswordResetCalls) != 1 {
			t.Fatalf("expected reset to be removed")
		}
		if q.SystemDeletePasswordResetCalls[0] != 1 {
			t.Fatalf("expected delete id 1 got %d", q.SystemDeletePasswordResetCalls[0])
		}
		if len(q.CreatePasswordResetForUserCalls) != 1 {
			t.Fatalf("expected one new reset, got %d", len(q.CreatePasswordResetForUserCalls))
		}
	})
}
