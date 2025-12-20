package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type commentPageQueries struct {
	db.Querier
	commentID  int32
	comment    *db.GetCommentByIdForUserRow
	threadInfo []*db.GetCommentsByIdsForUserWithThreadInfoRow
	threadRows []*db.GetCommentsByThreadIdForUserRow
}

func (q *commentPageQueries) GetCommentByIdForUser(_ context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
	if arg.ID != q.commentID {
		return nil, fmt.Errorf("unexpected comment id: %d", arg.ID)
	}
	return q.comment, nil
}

func (q *commentPageQueries) GetCommentsByIdsForUserWithThreadInfo(context.Context, db.GetCommentsByIdsForUserWithThreadInfoParams) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, error) {
	return q.threadInfo, nil
}

func (q *commentPageQueries) GetCommentsByThreadIdForUser(context.Context, db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	return q.threadRows, nil
}

func TestAdminCommentPage_UsesURLParam(t *testing.T) {
	commentID := 44
	threadID := 55
	topicID := 66
	queries := &commentPageQueries{
		commentID: int32(commentID),
		comment: &db.GetCommentByIdForUserRow{
			Idcomments:    int32(commentID),
			ForumthreadID: int32(threadID),
			UsersIdusers:  2,
			LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
			Written:       sql.NullTime{Time: time.Now(), Valid: true},
			Text:          sql.NullString{String: "body", Valid: true},
			Timezone:      sql.NullString{},
			DeletedAt:     sql.NullTime{},
			LastIndex:     sql.NullTime{},
			Username:      sql.NullString{String: "user", Valid: true},
			IsOwner:       true,
		},
		threadInfo: []*db.GetCommentsByIdsForUserWithThreadInfoRow{{
			Idcomments:         int32(commentID),
			ForumthreadID:      int32(threadID),
			UsersIdusers:       2,
			LanguageID:         sql.NullInt32{Int32: 1, Valid: true},
			Written:            sql.NullTime{Time: time.Now(), Valid: true},
			Text:               sql.NullString{String: "body", Valid: true},
			Timezone:           sql.NullString{},
			DeletedAt:          sql.NullTime{},
			LastIndex:          sql.NullTime{},
			Posterusername:     sql.NullString{String: "user", Valid: true},
			IsOwner:            true,
			Idforumthread:      sql.NullInt32{Int32: int32(threadID), Valid: true},
			Idforumtopic:       sql.NullInt32{Int32: int32(topicID), Valid: true},
			ForumtopicTitle:    sql.NullString{String: "topic", Valid: true},
			ThreadTitle:        sql.NullString{String: "thread", Valid: true},
			Idforumcategory:    sql.NullInt32{Int32: 1, Valid: true},
			ForumcategoryTitle: sql.NullString{String: "cat", Valid: true},
		}},
		threadRows: []*db.GetCommentsByThreadIdForUserRow{{
			Idcomments:     int32(commentID),
			ForumthreadID:  int32(threadID),
			UsersIdusers:   2,
			LanguageID:     sql.NullInt32{Int32: 1, Valid: true},
			Written:        sql.NullTime{Time: time.Now(), Valid: true},
			Text:           sql.NullString{String: "body", Valid: true},
			Timezone:       sql.NullString{},
			DeletedAt:      sql.NullTime{},
			LastIndex:      sql.NullTime{},
			Posterusername: sql.NullString{String: "user", Valid: true},
			IsOwner:        true,
		}},
	}

	req := httptest.NewRequest("GET", "/admin/comment/"+strconv.Itoa(commentID), nil)
	req = mux.SetURLVars(req, map[string]string{"comment": strconv.Itoa(commentID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminCommentPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
