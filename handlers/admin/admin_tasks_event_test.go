package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
)

type captureDLQ struct {
	lastError string
}

func (c *captureDLQ) Record(ctx context.Context, message string) error {
	c.lastError = message
	return nil
}

func setupTest(t *testing.T) (*db.QuerierStub, *eventbus.Bus, *notifications.Notifier, *captureDLQ, *sessions.CookieStore) {
	uid := int32(1)
	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
		return &db.SystemGetUserByIDRow{
			Idusers:  uid,
			Username: sql.NullString{String: "adminuser", Valid: true},
			Email:    sql.NullString{String: "admin@example.com", Valid: true},
		}, nil
	}
	qs.AdminListAdministratorEmailsReturns = []string{"root@example.com"}
	qs.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{
		Idusers: 99,
	}
	qs.GetPermissionsByUserIDFn = func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
		return []*db.GetPermissionsByUserIDRow{
			{Name: "admin", IsAdmin: true},
		}, nil
	}

	bus := eventbus.NewBus()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true
	cfg.AdminNotify = true
	cfg.EmailFrom = "test@example.com"
	n := notifications.New(notifications.WithQueries(qs), notifications.WithConfig(cfg))
	cdlq := &captureDLQ{}
	n.RegisterSync(bus, cdlq)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	return qs, bus, n, cdlq, store
}

func createRequest(ctx context.Context, method, url string, body string, sess *sessions.Session) *http.Request {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	_ = sess.Save(req, w)
	return req
}

func TestAddIPBanTaskEventData(t *testing.T) {
	qs, bus, _, cdlq, store := setupTest(t)
	uid := int32(1)

	qs.AdminInsertBannedIpFn = func(ctx context.Context, arg db.AdminInsertBannedIpParams) error {
		return nil
	}

	sess := testhelpers.Must(store.New(httptest.NewRequest("GET", "/", nil), "test"))
	sess.Values["UID"] = uid

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  uid,
		Path:    "/admin/ipban/add",
		Task:    addIPBanTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"admin"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"ip": {"192.168.1.1"}, "reason": {"Spam"}}
	req := createRequest(ctx, http.MethodPost, "http://example.com/admin/ipban/add", form.Encode(), sess)

	rr := httptest.NewRecorder()
	addIPBanTask.Action(rr, req)
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	if v, ok := evt.Data["IP"].(string); !ok || v != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %v", evt.Data["IP"])
	}

	// Verify Admin Notification logic
	found := false
	for _, call := range qs.SystemCreateNotificationCalls {
		if strings.Contains(call.Message.String, "adminuser") && strings.Contains(call.Message.String, "192.168.1.1") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Admin notification not found")
	}
}

func TestDeleteIPBanTaskEventData(t *testing.T) {
	qs, bus, _, cdlq, store := setupTest(t)
	uid := int32(1)

	qs.AdminCancelBannedIpFn = func(ctx context.Context, ip string) error {
		return nil
	}

	sess := testhelpers.Must(store.New(httptest.NewRequest("GET", "/", nil), "test"))
	sess.Values["UID"] = uid

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  uid,
		Path:    "/admin/ipban/delete",
		Task:    deleteIPBanTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"admin"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"ip": {"192.168.1.1"}}
	req := createRequest(ctx, http.MethodPost, "http://example.com/admin/ipban/delete", form.Encode(), sess)

	rr := httptest.NewRecorder()
	deleteIPBanTask.Action(rr, req)
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	if v, ok := evt.Data["IP"].(string); !ok || v != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %v", evt.Data["IP"])
	}

	found := false
	for _, call := range qs.SystemCreateNotificationCalls {
		if strings.Contains(call.Message.String, "adminuser") && strings.Contains(call.Message.String, "removed ban") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Admin notification not found")
	}
}

func TestAddAnnouncementTaskEventData(t *testing.T) {
	qs, bus, _, cdlq, store := setupTest(t)
	uid := int32(1)

	qs.AdminPromoteAnnouncementFn = func(ctx context.Context, id int32) error {
		return nil
	}

	sess := testhelpers.Must(store.New(httptest.NewRequest("GET", "/", nil), "test"))
	sess.Values["UID"] = uid

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  uid,
		Path:    "/admin/announcement/add",
		Task:    addAnnouncementTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"admin"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"news_id": {"100"}}
	req := createRequest(ctx, http.MethodPost, "http://example.com/admin/announcement/add", form.Encode(), sess)

	rr := httptest.NewRecorder()
	addAnnouncementTask.Action(rr, req)
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	if v, ok := evt.Data["NewsID"].(int); !ok || v != 100 {
		t.Errorf("expected NewsID 100, got %v", evt.Data["NewsID"])
	}

	found := false
	for _, call := range qs.SystemCreateNotificationCalls {
		msg := call.Message.String
		// Template announcement: "Announcement updated by {{.Item.Username}}"
		if strings.Contains(msg, "Announcement updated by adminuser") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Admin notification not found. Calls: %v", qs.SystemCreateNotificationCalls)
	}
}

func TestDeleteAnnouncementTaskEventData(t *testing.T) {
	qs, bus, _, cdlq, store := setupTest(t)
	uid := int32(1)

	qs.AdminDemoteAnnouncementFn = func(ctx context.Context, id int32) error {
		return nil
	}

	sess := testhelpers.Must(store.New(httptest.NewRequest("GET", "/", nil), "test"))
	sess.Values["UID"] = uid

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  uid,
		Path:    "/admin/announcement/delete",
		Task:    deleteAnnouncementTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"admin"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"id": {"101"}}
	req := createRequest(ctx, http.MethodPost, "http://example.com/admin/announcement/delete", form.Encode(), sess)

	rr := httptest.NewRecorder()
	deleteAnnouncementTask.Action(rr, req)
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	if v, ok := evt.Data["AnnouncementID"].(int); !ok || v != 101 {
		t.Errorf("expected AnnouncementID 101, got %v", evt.Data["AnnouncementID"])
	}

	found := false
	for _, call := range qs.SystemCreateNotificationCalls {
		msg := call.Message.String
		// Template announcement: "Announcement updated by {{.Item.Username}}"
		if strings.Contains(msg, "Announcement updated by adminuser") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Admin notification not found. Calls: %+v", qs.SystemCreateNotificationCalls)
	}
}

func TestUserPasswordResetTaskEventData(t *testing.T) {
	qs, bus, _, cdlq, store := setupTest(t)
	uid := int32(1)
	targetID := int32(2)

	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
		if idusers == uid {
			return &db.SystemGetUserByIDRow{
				Idusers:  uid,
				Username: sql.NullString{String: "adminuser", Valid: true},
				Email:    sql.NullString{String: "admin@example.com", Valid: true},
			}, nil
		}
		if idusers == targetID {
			return &db.SystemGetUserByIDRow{
				Idusers:  targetID,
				Username: sql.NullString{String: "targetuser", Valid: true},
				Email:    sql.NullString{String: "target@example.com", Valid: true},
			}, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.InsertPasswordFn = func(ctx context.Context, arg db.InsertPasswordParams) error {
		return nil
	}

	sess := testhelpers.Must(store.New(httptest.NewRequest("GET", "/", nil), "test"))
	sess.Values["UID"] = uid

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  uid,
		Path:    "/admin/user/2/password/reset",
		Task:    userForcePasswordChangeTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"admin"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := createRequest(ctx, http.MethodPost, "http://example.com/admin/user/2/password/reset", "", sess)
	req = mux.SetURLVars(req, map[string]string{"user": "2"})

	rr := httptest.NewRecorder()
	userForcePasswordChangeTask.Action(rr, req)
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	if v, ok := evt.Data["Username"].(string); !ok || v != "targetuser" {
		t.Errorf("expected Username 'targetuser', got %v", evt.Data["Username"])
	}

	found := false
	for _, call := range qs.SystemCreateNotificationCalls {
		if call.RecipientID == targetID {
			msg := call.Message.String
			if strings.Contains(msg, "reset your password") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("Target user notification not found")
	}
}

func TestServerShutdownTaskEventData(t *testing.T) {
	qs, bus, _, cdlq, store := setupTest(t)
	uid := int32(1)

	sess := testhelpers.Must(store.New(httptest.NewRequest("GET", "/", nil), "test"))
	sess.Values["UID"] = uid

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  uid,
		Path:    "/admin/shutdown",
		Task:    &ServerShutdownTask{},
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"admin"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := createRequest(ctx, http.MethodPost, "http://example.com/admin/shutdown", "", sess)

	rr := httptest.NewRecorder()
	(&ServerShutdownTask{}).Action(rr, req)
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	// ServerSummaryTask doesn't implement notification interfaces currently,
	// so no notifications are expected.
}
