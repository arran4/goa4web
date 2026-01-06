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
	"github.com/arran4/goa4web/internal/db"
)

type commentPageQueries struct {
	db.QuerierStub
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

// SystemCheckRoleGrant mocks the role check to prevent test panic/noise.
func (q *commentPageQueries) SystemCheckRoleGrant(context.Context, db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows // Default to no role grant found
}

// GetPermissionsByUserID mocks permission check to prevent test panic/noise.
func (q *commentPageQueries) GetPermissionsByUserID(context.Context, int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return nil, nil // Default to no permissions
}

// GetCommentsBySectionThreadIdForUser mocks retrieval for comment editing checks.
func (q *commentPageQueries) GetCommentsBySectionThreadIdForUser(context.Context, db.GetCommentsBySectionThreadIdForUserParams) ([]*db.GetCommentsBySectionThreadIdForUserRow, error) {
	return nil, nil
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
	cfg.TemplatesDir = "../../core/templates"
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

func TestAdminCommentPage_RendersCorrectTopicLink(t *testing.T) {

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
	cfg.TemplatesDir = "../../core/templates"
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminCommentPage(rr, req)
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
}
