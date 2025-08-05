package imagebbs

import (
	"context"
	"net/http/httptest"
	"regexp"
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

	boardRows := sqlmock.NewRows([]string{"idimageboard", "imageboard_idimageboard", "title", "description", "approval_required"}).
		AddRow(4, 3, "child", "sub", false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idimageboard, b.imageboard_idimageboard, b.title")).
		WithArgs(int32(0), int32(3), sqlmock.AnyArg(), int32(200), int32(0)).
		WillReturnRows(boardRows)

	postRows := sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "last_index", "username", "comments"}).
		AddRow(1, 1, 1, 3, time.Unix(0, 0), "desc", "/t", "/f", 10, true, nil, nil, "alice", 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.idimagepost, i.forumthread_id, i.users_idusers")).
		WithArgs(int32(0), int32(3), sqlmock.AnyArg(), int32(200), int32(0)).
		WillReturnRows(postRows)

	req := httptest.NewRequest("GET", "/imagebbs/board/3", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "3"})
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	BoardPage(rr, req)

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
