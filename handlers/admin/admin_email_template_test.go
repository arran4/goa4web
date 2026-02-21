package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	logProv "github.com/arran4/goa4web/internal/email/log"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func newEmailReg() *email.Registry {
	r := email.NewRegistry()
	logProv.Register(r)
	return r
}

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.EmailProvider = ""

		req := httptest.NewRequest("POST", "/admin/email/template", nil)
		reg := newEmailReg()
		p := testhelpers.Must(reg.ProviderFromConfig(cfg))
		qs := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(req.Context(), qs, cfg, common.WithEmailProvider(p), common.WithSilence(true))
		cd.UserID = 1
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(testTemplateTask)(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.EmailProvider = "log"

		qs := testhelpers.NewQuerierStub()
		qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
			Idusers:                1,
			Email:                  sql.NullString{String: "u@example.com", Valid: true},
			Username:               sql.NullString{String: "u", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		}
		qs.SystemGetTemplateOverrideReturns = ""
		// Override SystemGetTemplateOverrideFn to match behavior: return empty string if found (mocking template override)
		qs.SystemGetTemplateOverrideFn = func(_ context.Context, name string) (string, error) {
			if name == "updateEmail.gotxt" {
				return "", nil
			}
			return "", sql.ErrNoRows
		}

		req := httptest.NewRequest("POST", "/admin/email/template", nil)
		reg := newEmailReg()
		p := testhelpers.Must(reg.ProviderFromConfig(cfg))
		cd := common.NewCoreData(req.Context(), qs, cfg, common.WithEmailProvider(p), common.WithSilence(true))
		cd.UserID = 1
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(testTemplateTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(qs.InsertPendingEmailCalls) != 1 {
			t.Fatalf("expected pending email, got %d", len(qs.InsertPendingEmailCalls))
		}
		if pending := qs.InsertPendingEmailCalls[0]; !pending.ToUserID.Valid || pending.ToUserID.Int32 != 1 || pending.Body == "" || pending.DirectEmail {
			t.Fatalf("unexpected pending email: %#v", pending)
		}
	})
}

func TestAdminListUnsentPendingEmails(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		qs := testhelpers.NewQuerierStub()
		qs.AdminListUnsentPendingEmailsReturns = []*db.AdminListUnsentPendingEmailsRow{{
			ID:          1,
			ToUserID:    sql.NullInt32{Int32: 2, Valid: true},
			Body:        "b",
			ErrorCount:  0,
			CreatedAt:   time.Now(),
			DirectEmail: false,
		}}

		if _, err := qs.AdminListUnsentPendingEmails(context.Background(), db.AdminListUnsentPendingEmailsParams{}); err != nil {
			t.Fatalf("list: %v", err)
		}
	})
}

func TestRecentNotifications(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		qs := testhelpers.NewQuerierStub()
		qs.AdminListRecentNotificationsReturns = []*db.Notification{{
			ID:           1,
			UsersIdusers: 1,
			Link:         sql.NullString{String: "/l", Valid: true},
			Message:      sql.NullString{String: "m", Valid: true},
			CreatedAt:    time.Now(),
			ReadAt:       sql.NullTime{},
		}}

		if _, err := qs.AdminListRecentNotifications(context.Background(), 5); err != nil {
			t.Fatalf("recent: %v", err)
		}
	})
}

type recordAdminMail struct{ to []string }

func (r *recordAdminMail) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	r.to = append(r.to, to.Address)
	return nil
}

func (r *recordAdminMail) TestConfig(ctx context.Context) (string, error) { return "", nil }

func TestNotifyAdminsEnv(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.AdminEmails = "a@test.com,b@test.com"
		cfg.AdminNotify = true
		cfg.EmailEnabled = true
		cfg.EmailFrom = "from@example.com"

		os.Setenv(config.EnvAdminEmails, "a@test.com,b@test.com")
		defer os.Unsetenv(config.EnvAdminEmails)
		cfg = config.NewRuntimeConfig()
		origEmails := cfg.AdminEmails
		cfg.AdminEmails = "a@test.com,b@test.com"
		defer func() { cfg.AdminEmails = origEmails }()

		qs := testhelpers.NewQuerierStub()
		qs.SystemGetUserByEmailFn = func(_ context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
			if email == "a@test.com" {
				return &db.SystemGetUserByEmailRow{Idusers: 1, Email: "a@test.com", Username: sql.NullString{String: "a", Valid: true}}, nil
			}
			if email == "b@test.com" {
				return &db.SystemGetUserByEmailRow{Idusers: 2, Email: "b@test.com", Username: sql.NullString{String: "b", Valid: true}}, nil
			}
			return nil, sql.ErrNoRows
		}
		qs.SystemGetTemplateOverrideFn = func(_ context.Context, name string) (string, error) {
			if name == "adminNotificationEmailSubject.gotxt" {
				return "", nil
			}
			return "", sql.ErrNoRows
		}
		qs.ListSubscribersForPatternsFn = func(_ context.Context, arg db.ListSubscribersForPatternsParams) ([]int32, error) {
			var ret []int32
			for _, p := range arg.Patterns {
				if p == "notify:/admin/*" {
					ret = append(ret, 1, 2)
				}
			}
			return ret, nil
		}

		rec := &recordAdminMail{}
		n := notif.New(notif.WithSilence(true), notif.WithQueries(qs), notif.WithEmailProvider(rec), notif.WithConfig(cfg))
		n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
		if len(rec.to) != 0 {
			t.Fatalf("expected 0 direct mails, got %d", len(rec.to))
		}
		if len(qs.InsertPendingEmailCalls) != 2 {
			t.Fatalf("expected 2 pending emails, got %d", len(qs.InsertPendingEmailCalls))
		}
		for _, pending := range qs.InsertPendingEmailCalls {
			if pending.Body == "" || pending.DirectEmail {
				t.Fatalf("unexpected pending email: %#v", pending)
			}
		}
	})
}

func TestNotifyAdminsDisabled(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.AdminEmails = "a@test.com"
		cfg.AdminNotify = false
		cfg.EmailEnabled = true
		cfg.AdminEmails = "a@test.com"
		os.Setenv(config.EnvAdminNotify, "false")
		cfg.AdminEmails = "a@test.com"
		defer os.Unsetenv(config.EnvAdminEmails)
		defer os.Unsetenv(config.EnvAdminNotify)
		cfg.AdminEmails = "a@test.com"
		rec := &recordAdminMail{}
		n := notif.New(notif.WithSilence(true), notif.WithEmailProvider(rec), notif.WithConfig(cfg))
		n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
		if len(rec.to) != 0 {
			t.Fatalf("expected 0 mails, got %d", len(rec.to))
		}
	})
}
