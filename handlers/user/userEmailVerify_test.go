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

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type verifyEmailQueries struct {
	db.Querier
	code            string
	email           *db.UserEmail
	marked          []db.SystemMarkUserEmailVerifiedParams
	prioritySetArgs []db.SetNotificationPriorityForListerParams
	deletedArgs     []db.SystemDeleteUserEmailsByEmailExceptIDParams
	maxPriority     interface{}
}

func (q *verifyEmailQueries) GetUserEmailByCode(_ context.Context, code sql.NullString) (*db.UserEmail, error) {
	if code.String != q.code {
		return nil, fmt.Errorf("unexpected code: %s", code.String)
	}
	return q.email, nil
}

func (q *verifyEmailQueries) SystemMarkUserEmailVerified(_ context.Context, arg db.SystemMarkUserEmailVerifiedParams) error {
	q.marked = append(q.marked, arg)
	return nil
}

func (q *verifyEmailQueries) GetMaxNotificationPriority(_ context.Context, userID int32) (interface{}, error) {
	return q.maxPriority, nil
}

func (q *verifyEmailQueries) SetNotificationPriorityForLister(_ context.Context, arg db.SetNotificationPriorityForListerParams) error {
	q.prioritySetArgs = append(q.prioritySetArgs, arg)
	return nil
}

func (q *verifyEmailQueries) SystemDeleteUserEmailsByEmailExceptID(_ context.Context, arg db.SystemDeleteUserEmailsByEmailExceptIDParams) error {
	q.deletedArgs = append(q.deletedArgs, arg)
	return nil
}

func TestUserEmailVerifyCodePage_Invalid(t *testing.T) {
	code := "abc"
	q := &verifyEmailQueries{
		code:  code,
		email: &db.UserEmail{ID: 1, UserID: 1, Email: "e@example.com", LastVerificationCode: sql.NullString{String: code, Valid: true}},
	}

	store := sessions.NewCookieStore([]byte("test"))
	sess := sessions.NewSession(store, "test")
	sess.Values = map[interface{}]interface{}{"UID": int32(2)}
	core.Store = store
	core.SessionName = "test"

	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := httptest.NewRequest(http.MethodGet, "/usr/email/verify?code="+code, nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	userEmailVerifyCodePage(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestUserEmailVerifyCodePage_Success(t *testing.T) {
	code := "xyz"
	q := &verifyEmailQueries{
		code:        code,
		email:       &db.UserEmail{ID: 1, UserID: 1, Email: "e@example.com", LastVerificationCode: sql.NullString{String: code, Valid: true}},
		maxPriority: int64(0),
	}

	store := sessions.NewCookieStore([]byte("test"))
	sess := sessions.NewSession(store, "test")
	sess.Values = map[interface{}]interface{}{"UID": int32(1)}
	core.Store = store
	core.SessionName = "test"

	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"code": {code}}
	req := httptest.NewRequest(http.MethodPost, "/usr/email/verify", strings.NewReader(form.Encode())).WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	userEmailVerifyCodePage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/usr/email" {
		t.Fatalf("location=%q", loc)
	}
	if len(q.marked) != 1 || q.marked[0].ID != 1 {
		t.Fatalf("unexpected mark args: %#v", q.marked)
	}
	if len(q.prioritySetArgs) != 1 {
		t.Fatalf("unexpected priority updates: %#v", q.prioritySetArgs)
	}
	if arg := q.prioritySetArgs[0]; arg.ListerID != 1 || arg.NotificationPriority != 1 || arg.ID != 1 {
		t.Fatalf("unexpected priority args: %#v", arg)
	}
	if len(q.deletedArgs) != 1 {
		t.Fatalf("unexpected delete args: %#v", q.deletedArgs)
	}
	if arg := q.deletedArgs[0]; arg.Email != "e@example.com" || arg.ID != 1 {
		t.Fatalf("unexpected delete args: %#v", arg)
	}
}
