package admin

import (
	"context"
	"database/sql"
	"fmt"
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

type emailTemplateQueries struct {
	db.Querier
	userID            int32
	user              *db.SystemGetUserByIDRow
	templateOverrides map[string]string
	pendingEmails     []db.InsertPendingEmailParams
}

func (q *emailTemplateQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *emailTemplateQueries) SystemGetTemplateOverride(_ context.Context, name string) (string, error) {
	body, ok := q.templateOverrides[name]
	if !ok {
		return "", sql.ErrNoRows
	}
	return body, nil
}

func (q *emailTemplateQueries) InsertPendingEmail(_ context.Context, arg db.InsertPendingEmailParams) error {
	q.pendingEmails = append(q.pendingEmails, arg)
	return nil
}

type pendingEmailQueries struct {
	db.Querier
	rows []*db.AdminListUnsentPendingEmailsRow
}

func (q *pendingEmailQueries) AdminListUnsentPendingEmails(context.Context, db.AdminListUnsentPendingEmailsParams) ([]*db.AdminListUnsentPendingEmailsRow, error) {
	return q.rows, nil
}

type recentNotificationsQueries struct {
	db.Querier
	limit         int32
	notifications []*db.Notification
}

func (q *recentNotificationsQueries) AdminListRecentNotifications(_ context.Context, limit int32) ([]*db.Notification, error) {
	if q.limit != 0 && limit != q.limit {
		return nil, fmt.Errorf("unexpected limit: %d", limit)
	}
	return q.notifications, nil
}

type notifyAdminsQueries struct {
	db.Querier
	usersByEmail      map[string]*db.SystemGetUserByEmailRow
	templateOverrides map[string]string
	pendingEmails     []db.InsertPendingEmailParams
}

func (q *notifyAdminsQueries) SystemGetUserByEmail(_ context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
	if user, ok := q.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, sql.ErrNoRows
}

func (q *notifyAdminsQueries) SystemGetTemplateOverride(_ context.Context, name string) (string, error) {
	body, ok := q.templateOverrides[name]
	if !ok {
		return "", sql.ErrNoRows
	}
	return body, nil
}

func (q *notifyAdminsQueries) InsertPendingEmail(_ context.Context, arg db.InsertPendingEmailParams) error {
	q.pendingEmails = append(q.pendingEmails, arg)
	return nil
}

func newEmailReg() *email.Registry {
	r := email.NewRegistry()
	logProv.Register(r)
	return r
}

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = ""

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	reg := newEmailReg()
	p := testhelpers.Must(reg.ProviderFromConfig(cfg))
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithEmailProvider(p))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(testTemplateTask)(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = "log"

	queries := &emailTemplateQueries{
		userID: 1,
		user: &db.SystemGetUserByIDRow{
			Idusers:                1,
			Email:                  sql.NullString{String: "u@example.com", Valid: true},
			Username:               sql.NullString{String: "u", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
		templateOverrides: map[string]string{
			"updateEmail.gotxt": "",
		},
	}

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	reg := newEmailReg()
	p := testhelpers.Must(reg.ProviderFromConfig(cfg))
	cd := common.NewCoreData(req.Context(), queries, cfg, common.WithEmailProvider(p))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(testTemplateTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(queries.pendingEmails) != 1 {
		t.Fatalf("expected pending email, got %d", len(queries.pendingEmails))
	}
	if pending := queries.pendingEmails[0]; !pending.ToUserID.Valid || pending.ToUserID.Int32 != 1 || pending.Body == "" || pending.DirectEmail {
		t.Fatalf("unexpected pending email: %#v", pending)
	}
}

func TestAdminListUnsentPendingEmails(t *testing.T) {
	q := &pendingEmailQueries{
		rows: []*db.AdminListUnsentPendingEmailsRow{{
			ID:          1,
			ToUserID:    sql.NullInt32{Int32: 2, Valid: true},
			Body:        "b",
			ErrorCount:  0,
			CreatedAt:   time.Now(),
			DirectEmail: false,
		}},
	}
	if _, err := q.AdminListUnsentPendingEmails(context.Background(), db.AdminListUnsentPendingEmailsParams{}); err != nil {
		t.Fatalf("list: %v", err)
	}
}

func TestRecentNotifications(t *testing.T) {
	q := &recentNotificationsQueries{
		limit: 5,
		notifications: []*db.Notification{{
			ID:           1,
			UsersIdusers: 1,
			Link:         sql.NullString{String: "/l", Valid: true},
			Message:      sql.NullString{String: "m", Valid: true},
			CreatedAt:    time.Now(),
			ReadAt:       sql.NullTime{},
		}},
	}
	if _, err := q.AdminListRecentNotifications(context.Background(), 5); err != nil {
		t.Fatalf("recent: %v", err)
	}
}

type recordAdminMail struct{ to []string }

func (r *recordAdminMail) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	r.to = append(r.to, to.Address)
	return nil
}

func (r *recordAdminMail) TestConfig(ctx context.Context) error { return nil }

func TestNotifyAdminsEnv(t *testing.T) {
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
	q := &notifyAdminsQueries{
		usersByEmail: map[string]*db.SystemGetUserByEmailRow{
			"a@test.com": {Idusers: 1, Email: "a@test.com", Username: sql.NullString{String: "a", Valid: true}},
			"b@test.com": {Idusers: 2, Email: "b@test.com", Username: sql.NullString{String: "b", Valid: true}},
		},
		templateOverrides: map[string]string{
			"adminNotificationEmailSubject.gotxt": "",
		},
	}

	rec := &recordAdminMail{}
	n := notif.New(notif.WithQueries(q), notif.WithEmailProvider(rec), notif.WithConfig(cfg))
	n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 direct mails, got %d", len(rec.to))
	}
	if len(q.pendingEmails) != 2 {
		t.Fatalf("expected 2 pending emails, got %d", len(q.pendingEmails))
	}
	for _, pending := range q.pendingEmails {
		if pending.Body == "" || pending.DirectEmail {
			t.Fatalf("unexpected pending email: %#v", pending)
		}
	}
}

func TestNotifyAdminsDisabled(t *testing.T) {
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
	n := notif.New(notif.WithEmailProvider(rec), notif.WithConfig(cfg))
	n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 mails, got %d", len(rec.to))
	}
}
