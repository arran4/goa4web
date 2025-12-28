package forum

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestCustomForumIndexWriteReply(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

	mock.ExpectQuery(`WITH role_ids AS \( SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = \? UNION SELECT id FROM roles WHERE name = 'anyone' \) SELECT 1 FROM grants`).
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("expected write reply item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexMarkReadLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 7

	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(7)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT item, item_id, user_id, label, invert\\s+FROM content_private_labels").
		WithArgs("thread", int32(3), int32(7)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}).
			AddRow("thread", 3, 7, "unread", false))
	mock.ExpectQuery(`WITH role_ids AS \( SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = \? UNION SELECT id FROM roles WHERE name = 'anyone' \) SELECT 1 FROM grants`).
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomForumIndex(cd, req.WithContext(ctx))

	for _, name := range []string{"Mark as read", "Mark as read and go back", "Go to topic"} {
		if !common.ContainsItem(cd.CustomIndexItems, name) {
			t.Errorf("expected %s item", name)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexHidesMarkReadWhenClear(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 7

	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(7)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT item, item_id, user_id, label, invert\\s+FROM content_private_labels").
		WithArgs("thread", int32(3), int32(7)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}).
			AddRow("thread", 3, 7, "unread", true).
			AddRow("thread", 3, 7, "new", true))
	mock.ExpectQuery(`WITH role_ids AS \( SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = \? UNION SELECT id FROM roles WHERE name = 'anyone' \) SELECT 1 FROM grants`).
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomForumIndex(cd, req.WithContext(ctx))

	for _, name := range []string{"Mark as read", "Mark as read and go back"} {
		if common.ContainsItem(cd.CustomIndexItems, name) {
			t.Errorf("unexpected %s item", name)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexWriteReplyDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1

	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`WITH role_ids AS \( SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = \? UNION SELECT id FROM roles WHERE name = 'anyone' \) SELECT 1 FROM grants`).
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	CustomForumIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("unexpected write reply item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexCreateThread(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1

	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`WITH role_ids AS \( SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = \? UNION SELECT id FROM roles WHERE name = 'anyone' \) SELECT 1 FROM grants`).
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("expected create thread item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexAdminEditLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	ctx := req.Context()
	cd := common.NewCoreData(ctx, &db.QuerierStub{}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.UserID = 1
	cd.AdminMode = true
	cd.LoadSelectionsFromRequest(req)

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Admin Edit Topic") {
		t.Errorf("expected admin edit link")
	}
}

func TestCustomForumIndexCreateThreadDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1

	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`WITH role_ids AS \( SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = \? UNION SELECT id FROM roles WHERE name = 'anyone' \) SELECT 1 FROM grants`).
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	CustomForumIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("unexpected create thread item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexSubscribeLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1

	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT id, pattern, method FROM subscriptions").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "pattern", "method"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Subscribe To Topic") {
		t.Errorf("expected subscribe item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexUnsubscribeLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1

	pattern := topicSubscriptionPattern(2)
	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT id, pattern, method FROM subscriptions").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "pattern", "method"}).AddRow(1, pattern, "internal"))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Unsubscribe From Topic") {
		t.Errorf("expected unsubscribe item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
