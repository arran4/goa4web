package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathAdminCommentPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		commentID := 44
		threadID := 55
		topicID := 66

		queries := testhelpers.NewQuerierStub()
		queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
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
		}
		queries.GetCommentsByIdsForUserWithThreadInfoReturns = []*db.GetCommentsByIdsForUserWithThreadInfoRow{{
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
		}}
		queries.GetCommentsByThreadIdForUserReturns = []*db.GetCommentsByThreadIdForUserRow{{
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
		}}
		queries.SystemCheckRoleGrantErr = sql.ErrNoRows

		req := httptest.NewRequest("GET", "/admin/comment/"+strconv.Itoa(commentID), nil)
		req = mux.SetURLVars(req, map[string]string{"comment": strconv.Itoa(commentID)})
		cfg := config.NewRuntimeConfig()
		cd := common.NewCoreData(req.Context(), queries, cfg)
		cd.LoadSelectionsFromRequest(req)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(&AdminCommentTask{})(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}

		body := rr.Body.String()
		// Check for malformed URL
		if strings.Contains(body, "topic/{") {
			t.Errorf("Found malformed topic link in body: ...%s...", body[strings.Index(body, "topic/"):strings.Index(body, "topic/")+20])
		}

		expectedLink := fmt.Sprintf("/forum/topic/%d", topicID)
		if !strings.Contains(body, expectedLink) {
			t.Errorf("Expected link %s not found in body", expectedLink)
		}

		expectedThreadLink := fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, threadID)
		if !strings.Contains(body, expectedThreadLink) {
			t.Errorf("Expected link %s not found in body", expectedThreadLink)
		}

		expectedAdminLink := fmt.Sprintf("/admin/forum/topic/%d", topicID)
		if !strings.Contains(body, expectedAdminLink) {
			t.Errorf("Expected admin link %s not found in body", expectedAdminLink)
		}
	})
}
