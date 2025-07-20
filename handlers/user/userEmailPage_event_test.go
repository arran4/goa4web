package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/gorilla/sessions"
)

func TestAddEmailTaskEventData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("SELECT id, user_id, email").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO user_emails").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT u.idusers").WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).
			AddRow(1, nil, "alice"))

	store := sessions.NewCookieStore([]byte("test"))
	sess := sessions.NewSession(store, "test")
	sess.Values = map[interface{}]interface{}{"UID": int32(1)}
	core.Store = store
	core.SessionName = "test"

	evt := &eventbus.Event{Data: map[string]any{}}
	ctx := context.WithValue(context.Background(), consts.KeyQueries, q)
	cd := common.NewCoreData(ctx, q, common.WithSession(sess), common.WithEvent(evt))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := httptest.NewRequest("POST", "http://example.com/usr/email", strings.NewReader("new_email=a@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	addEmailTask.Action(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["URL"]; !ok {
		t.Fatalf("missing URL event data: %+v", evt.Data)
	}
	if evt.Data["Username"] != "alice" {
		t.Fatalf("username not set: %+v", evt.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
