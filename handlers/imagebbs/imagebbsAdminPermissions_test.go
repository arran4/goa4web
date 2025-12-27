package imagebbs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequireImagebbsGrantWithBoard(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 99

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM grants g")).
		WithArgs(cd.UserID, "imagebbs", sql.NullString{String: "board", Valid: true}, imagebbsApproveAction, sql.NullInt32{Int32: 7, Valid: true}, sql.NullInt32{Int32: cd.UserID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := httptest.NewRequest("GET", "/admin/imagebbs/board/7", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"board": "7"})

	if !requireImagebbsGrant(imagebbsApproveAction)(req, &mux.RouteMatch{}) {
		t.Fatalf("expected matcher to allow request with grant")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRequireImagebbsGrantWithPost(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 42

	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.idimagepost")).
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "timezone", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "last_index", "username", "comments"}).
			AddRow(5, 0, 0, sql.NullInt32{Int32: 3, Valid: true}, nil, nil, nil, nil, nil, 0, false, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM grants g")).
		WithArgs(cd.UserID, "imagebbs", sql.NullString{String: "board", Valid: true}, imagebbsApproveAction, sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: cd.UserID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := httptest.NewRequest("GET", "/admin/imagebbs/approve/5", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "5"})

	if !requireImagebbsGrant(imagebbsApproveAction)(req, &mux.RouteMatch{}) {
		t.Fatalf("expected matcher to allow request with post-derived grant")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestApprovePostTaskDeniesWithoutGrant(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 77

	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.idimagepost")).
		WithArgs(int32(12)).
		WillReturnRows(sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "timezone", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "last_index", "username", "comments"}).
			AddRow(12, 0, 0, sql.NullInt32{Int32: 9, Valid: true}, nil, nil, nil, nil, nil, 0, false, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM grants g")).
		WithArgs(cd.UserID, "imagebbs", sql.NullString{String: "board", Valid: true}, imagebbsApproveAction, sql.NullInt32{Int32: 9, Valid: true}, sql.NullInt32{Int32: cd.UserID, Valid: true}).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("POST", "/admin/imagebbs/approve/12", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "12"})

	if _, ok := approvePostTask.Action(httptest.NewRecorder(), req).(http.HandlerFunc); !ok {
		t.Fatalf("expected forbidden handler when grant is missing")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestApprovePostTaskAllowsWithGrant(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 15

	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.idimagepost")).
		WithArgs(int32(4)).
		WillReturnRows(sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "timezone", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "last_index", "username", "comments"}).
			AddRow(4, 0, 0, sql.NullInt32{Int32: 2, Valid: true}, nil, nil, nil, nil, nil, 0, false, nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM grants g")).
		WithArgs(cd.UserID, "imagebbs", sql.NullString{String: "board", Valid: true}, imagebbsApproveAction, sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: cd.UserID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE imagepost SET approved")).
		WithArgs(int32(4)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/admin/imagebbs/approve/4", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "4"})

	if result := approvePostTask.Action(httptest.NewRecorder(), req); result != nil {
		t.Fatalf("unexpected result %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
