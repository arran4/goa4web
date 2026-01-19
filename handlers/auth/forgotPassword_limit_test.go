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
)

func TestForgotPasswordRateLimit(t *testing.T) {
	now := time.Now()
	q := &fakeForgotPasswordQueries{
		loginRow: &db.SystemGetLoginRow{
			Idusers: 1,
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		},
		verifiedEmails: []string{"a@test.com"},
		userByEmail: &db.SystemGetUserByEmailRow{
			Idusers: 1,
			Email:   "a@test.com",
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		},
		pending: &db.PendingPassword{
			ID:               1,
			UserID:           1,
			Passwd:           sql.NullString{String: "hash", Valid: true},
			PasswdAlgorithm:  sql.NullString{String: "alg", Valid: true},
			VerificationCode: "code",
			CreatedAt:        now,
			VerifiedAt:       sql.NullTime{},
		},
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
	if len(q.createdResets) != 0 {
		t.Fatalf("expected no reset inserts when rate limited, got %d", len(q.createdResets))
	}
	if q.deleteID != 0 {
		t.Fatalf("unexpected delete of reset id %d", q.deleteID)
	}
}

func TestForgotPasswordReplaceOld(t *testing.T) {
	q := &fakeForgotPasswordQueries{
		loginRow: &db.SystemGetLoginRow{
			Idusers: 1,
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		},
		verifiedEmails: []string{"a@test.com"},
		userByEmail: &db.SystemGetUserByEmailRow{
			Idusers: 1,
			Email:   "a@test.com",
			Username: sql.NullString{
				String: "u",
				Valid:  true,
			},
		},
		pending: &db.PendingPassword{
			ID:               1,
			UserID:           1,
			Passwd:           sql.NullString{String: "hash", Valid: true},
			PasswdAlgorithm:  sql.NullString{String: "alg", Valid: true},
			VerificationCode: "code",
			CreatedAt:        time.Now().Add(-25 * time.Hour),
			VerifiedAt:       sql.NullTime{},
		},
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
	if q.deleteID != q.pending.ID {
		t.Fatalf("expected reset %d to be removed got %d", q.pending.ID, q.deleteID)
	}
	if len(q.createdResets) != 1 {
		t.Fatalf("expected one new reset, got %d", len(q.createdResets))
	}
}
