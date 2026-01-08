package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminCommentPage_UsesURLParam(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	commentID := 44
	threadID := 55
	topicID := 66

	rows1 := sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "username", "is_owner"}).
		AddRow(commentID, threadID, 2, 1, time.Now(), "body", nil, nil, nil, "user", true)
	mock.ExpectQuery("SELECT").WillReturnRows(rows1)

	rows2 := sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "posterusername", "is_owner", "idforumthread", "idforumtopic", "forumtopic_title", "thread_title", "idforumcategory", "forumcategory_title", "handler"}).
		AddRow(commentID, threadID, 2, 1, time.Now(), "body", nil, nil, nil, "user", true, threadID, topicID, "topic", "thread", 1, "cat", "forum")
	mock.ExpectQuery("SELECT").WillReturnRows(rows2)

	rows3 := sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "posterusername", "is_owner"}).
		AddRow(commentID, threadID, 2, 1, time.Now(), "body", nil, nil, nil, "user", true)
	mock.ExpectQuery("SELECT").WillReturnRows(rows3)

	req := httptest.NewRequest("GET", "/admin/comment/"+strconv.Itoa(commentID), nil)
	req = mux.SetURLVars(req, map[string]string{"comment": strconv.Itoa(commentID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminCommentPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
