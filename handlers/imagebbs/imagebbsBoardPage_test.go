package imagebbs

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestBoardPageRendersSubBoards(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	boardRows := sqlmock.NewRows([]string{"idimageboard", "imageboard_idimageboard", "title", "description", "approval_required", "deleted_at"}).
		AddRow(4, 3, "child", "sub", false, nil)
	mock.ExpectQuery("(?s)WITH role_ids AS .*SELECT b.idimageboard.*").
		WithArgs(int32(0), sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: 3, Valid: true}, sqlmock.AnyArg(), int32(200), int32(0)).
		WillReturnRows(boardRows)

	postRows := sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "timezone", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "last_index", "username", "comments"}).
		AddRow(1, 1, 1, 3, time.Unix(0, 0), time.Local.String(), "desc", "/t", "/f", 10, true, nil, nil, "alice", 0)
	mock.ExpectQuery("(?s)WITH role_ids AS .*SELECT i.idimagepost.*").
		WithArgs(int32(0), sql.NullInt32{Int32: 3, Valid: true}, sqlmock.AnyArg(), int32(200), int32(0)).
		WillReturnRows(postRows)

	// HasAdminRole (called by CanPostImage in template)
	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").WithArgs(int32(0)).WillReturnError(sql.ErrNoRows)
	// HasGrant (fallback)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(0), "imagebbs", sql.NullString{String: "board", Valid: true}, "post", sql.NullInt32{Int32: 3, Valid: true}, sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/imagebbs/board/3", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "3"})
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ImagebbsBoardPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Sub-Boards") {
		t.Fatalf("expected sub boards in output: %s", body)
	}
	if !strings.Contains(body, "Pictures:") {
		t.Fatalf("expected pictures in output: %s", body)
	}
}
