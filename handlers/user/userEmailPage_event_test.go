package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/gorilla/sessions"
)

type emailEventQueries struct {
	db.Querier
	userID   int32
	user     *db.SystemGetUserByIDRow
	inserted []db.InsertUserEmailParams
}

func (q *emailEventQueries) GetUserEmailByEmail(context.Context, string) (*db.UserEmail, error) {
	return nil, sql.ErrNoRows
}

func (q *emailEventQueries) InsertUserEmail(_ context.Context, arg db.InsertUserEmailParams) error {
	q.inserted = append(q.inserted, arg)
	return nil
}

func (q *emailEventQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

type verifyEventQueries struct {
	db.Querier
	code        string
	email       *db.UserEmail
	maxPriority interface{}
	marked      []db.SystemMarkUserEmailVerifiedParams
	priority    []db.SetNotificationPriorityForListerParams
	deleted     []db.SystemDeleteUserEmailsByEmailExceptIDParams
}

func (q *verifyEventQueries) GetUserEmailByCode(_ context.Context, code sql.NullString) (*db.UserEmail, error) {
	if code.String != q.code {
		return nil, fmt.Errorf("unexpected code: %s", code.String)
	}
	return q.email, nil
}

func (q *verifyEventQueries) SystemMarkUserEmailVerified(_ context.Context, arg db.SystemMarkUserEmailVerifiedParams) error {
	q.marked = append(q.marked, arg)
	return nil
}

func (q *verifyEventQueries) GetMaxNotificationPriority(context.Context, int32) (interface{}, error) {
	return q.maxPriority, nil
}

func (q *verifyEventQueries) SetNotificationPriorityForLister(_ context.Context, arg db.SetNotificationPriorityForListerParams) error {
	q.priority = append(q.priority, arg)
	return nil
}

func (q *verifyEventQueries) SystemDeleteUserEmailsByEmailExceptID(_ context.Context, arg db.SystemDeleteUserEmailsByEmailExceptIDParams) error {
	q.deleted = append(q.deleted, arg)
	return nil
}

type resendVerificationQueries struct {
	db.Querier
	emailID int32
	email   *db.UserEmail
	user    *db.SystemGetUserByIDRow
	updated []db.SetVerificationCodeForListerParams
}

func (q *resendVerificationQueries) GetUserEmailByID(_ context.Context, id int32) (*db.UserEmail, error) {
	if id != q.emailID {
		return nil, fmt.Errorf("unexpected email id: %d", id)
	}
	return q.email, nil
}

func (q *resendVerificationQueries) SetVerificationCodeForLister(_ context.Context, arg db.SetVerificationCodeForListerParams) error {
	q.updated = append(q.updated, arg)
	return nil
}

func (q *resendVerificationQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if q.user != nil && id != q.user.Idusers {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func TestAddEmailTaskEventData(t *testing.T) {
	q := &emailEventQueries{
		userID: 1,
		user: &db.SystemGetUserByIDRow{
			Idusers:                1,
			Username:               sql.NullString{String: "alice", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}

	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt))
	cd.UserID = 1
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	addEmailTask.codeGenerator = func() (string, error) { return "deadbeef", nil }
	defer func() { addEmailTask.codeGenerator = nil }()

	req := httptest.NewRequest("POST", "http://example.com/usr/email", strings.NewReader("new_email=a@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(addEmailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["URL"]; !ok {
		t.Fatalf("missing URL event data: %+v", evt.Data)
	}
	if evt.Data["VerificationCode"] != "deadbeef" {
		t.Fatalf("verification code not set: %+v", evt.Data)
	}
	if evt.Data["Token"] != "deadbeef" {
		t.Fatalf("token not set: %+v", evt.Data)
	}
	if evt.Data["Username"] != "alice" {
		t.Fatalf("username not set: %+v", evt.Data)
	}
}

func TestVerifyRemovesDuplicates(t *testing.T) {
	q := &verifyEventQueries{
		code: "code",
		email: &db.UserEmail{
			ID:                    1,
			UserID:                1,
			Email:                 "a@example.com",
			VerifiedAt:            sql.NullTime{},
			LastVerificationCode:  sql.NullString{String: "code", Valid: true},
			VerificationExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
			NotificationPriority:  0,
		},
		maxPriority: int64(0),
	}

	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt))
	cd.UserID = 1

	addEmailTask.codeGenerator = func() (string, error) { return "deadbeef", nil }
	defer func() { addEmailTask.codeGenerator = nil }()

	sess.Values = map[interface{}]interface{}{"UID": int32(1)}
	core.Store = store
	core.SessionName = "test"

	form := url.Values{"code": {"code"}}
	req := httptest.NewRequest(http.MethodPost, "/usr/email/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	userEmailVerifyCodePage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(q.marked) != 1 || q.marked[0].ID != 1 {
		t.Fatalf("unexpected mark args: %#v", q.marked)
	}
	if len(q.priority) != 1 {
		t.Fatalf("unexpected priority args: %#v", q.priority)
	}
	if arg := q.priority[0]; arg.ListerID != 1 || arg.NotificationPriority != 1 || arg.ID != 1 {
		t.Fatalf("unexpected priority args: %#v", arg)
	}
	if len(q.deleted) != 1 {
		t.Fatalf("unexpected delete args: %#v", q.deleted)
	}
	if arg := q.deleted[0]; arg.Email != "a@example.com" || arg.ID != 1 {
		t.Fatalf("unexpected delete args: %#v", arg)
	}
}

func TestResendVerificationEmailTaskEventData(t *testing.T) {
	q := &resendVerificationQueries{
		emailID: 1,
		email:   &db.UserEmail{ID: 1, UserID: 1, Email: "a@example.com"},
		user:    &db.SystemGetUserByIDRow{Idusers: 1, Username: sql.NullString{String: "alice", Valid: true}},
	}

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"

	req := httptest.NewRequest("POST", "http://example.com/usr/email/resend", strings.NewReader("id=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	addEmailTask.codeGenerator = func() (string, error) { return "deadbeef", nil }
	defer func() { addEmailTask.codeGenerator = nil }()

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(resendVerificationEmailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["page"]; !ok {
		t.Fatalf("missing page event data: %+v", evt.Data)
	}
	if _, ok := evt.Data["email"]; !ok {
		t.Fatalf("missing email event data: %+v", evt.Data)
	}
	if evt.Data["VerificationCode"] != "deadbeef" {
		t.Fatalf("verification code missing: %+v", evt.Data)
	}
	if evt.Data["Token"] != "deadbeef" {
		t.Fatalf("token missing: %+v", evt.Data)
	}
	if len(q.updated) != 1 {
		t.Fatalf("expected verification update, got %d", len(q.updated))
	}
	if arg := q.updated[0]; arg.ListerID != 1 || arg.ID != 1 {
		t.Fatalf("unexpected verification args: %#v", arg)
	}
}
