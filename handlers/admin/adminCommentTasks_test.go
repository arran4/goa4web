package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func setupCommentTest(t *testing.T, commentID int, body url.Values, queries db.Querier) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	var reader *strings.Reader
	if body != nil {
		reader = strings.NewReader(body.Encode())
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest("POST", "/admin/comment/"+strconv.Itoa(commentID), reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req = mux.SetURLVars(req, map[string]string{"comment": strconv.Itoa(commentID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return httptest.NewRecorder(), req
}

func TestHappyPathDeleteCommentTask(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments:    15,
		ForumthreadID: 2,
		UsersIdusers:  3,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Written:       sql.NullTime{Time: time.Now(), Valid: true},
		Text:          sql.NullString{String: "body", Valid: true},
		Timezone:      sql.NullString{},
		DeletedAt:     sql.NullTime{},
		LastIndex:     sql.NullTime{},
		Username:      sql.NullString{String: "user", Valid: true},
		IsOwner:       true,
	}

	rr, req := setupCommentTest(t, 15, nil, queries)
	if err, ok := deleteCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.AdminScrubCommentCalls) != 1 || queries.AdminScrubCommentCalls[0].Idcomments != 15 {
		t.Fatalf("unexpected scrub args: %#v", queries.AdminScrubCommentCalls)
	}
}

func TestHappyPathEditCommentTask(t *testing.T) {
	body := url.Values{"replytext": {"updated"}}
	queries := testhelpers.NewQuerierStub()
	queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments:    22,
		ForumthreadID: 2,
		UsersIdusers:  3,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Written:       sql.NullTime{Time: time.Now(), Valid: true},
		Text:          sql.NullString{String: "body", Valid: true},
		Timezone:      sql.NullString{},
		DeletedAt:     sql.NullTime{},
		LastIndex:     sql.NullTime{},
		Username:      sql.NullString{String: "user", Valid: true},
		IsOwner:       true,
	}

	rr, req := setupCommentTest(t, 22, body, queries)
	if err, ok := editCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.AdminScrubCommentCalls) != 1 || queries.AdminScrubCommentCalls[0].Idcomments != 22 || queries.AdminScrubCommentCalls[0].Text.String != "updated" {
		t.Fatalf("unexpected scrub args: %#v", queries.AdminScrubCommentCalls)
	}
}

func TestHappyPathDeactivateCommentTask(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.AdminIsCommentDeactivatedFn = func(_ context.Context, id int32) (bool, error) {
		if id != 33 {
			return false, sql.ErrNoRows
		}
		return false, nil
	}
	queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments:    33,
		ForumthreadID: 2,
		UsersIdusers:  3,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Written:       sql.NullTime{Time: time.Now(), Valid: true},
		Text:          sql.NullString{String: "body", Valid: true},
		Timezone:      sql.NullString{},
		DeletedAt:     sql.NullTime{},
		LastIndex:     sql.NullTime{},
		Username:      sql.NullString{String: "user", Valid: true},
		IsOwner:       true,
	}

	rr, req := setupCommentTest(t, 33, nil, queries)
	if err, ok := deactivateCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.AdminArchiveCommentCalls) != 1 || queries.AdminArchiveCommentCalls[0].Idcomments != 33 {
		t.Fatalf("unexpected archive args: %#v", queries.AdminArchiveCommentCalls)
	}
	if len(queries.AdminScrubCommentCalls) != 1 || queries.AdminScrubCommentCalls[0].Idcomments != 33 {
		t.Fatalf("unexpected scrub args: %#v", queries.AdminScrubCommentCalls)
	}
}

func TestHappyPathRestoreCommentTask(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.AdminIsCommentDeactivatedFn = func(_ context.Context, id int32) (bool, error) {
		if id != 44 {
			return false, sql.ErrNoRows
		}
		return true, nil
	}
	queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments:    44,
		ForumthreadID: 2,
		UsersIdusers:  3,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Written:       sql.NullTime{Time: time.Now(), Valid: true},
		Text:          sql.NullString{String: "", Valid: true},
		Timezone:      sql.NullString{},
		DeletedAt:     sql.NullTime{},
		LastIndex:     sql.NullTime{},
		Username:      sql.NullString{String: "user", Valid: true},
		IsOwner:       true,
	}
	queries.AdminListDeactivatedCommentsReturns = []*db.AdminListDeactivatedCommentsRow{{
		Idcomments: 44,
		Text:       sql.NullString{String: "body", Valid: true},
	}}
	var restoredIDs []int32
	queries.AdminMarkCommentRestoredFn = func(_ context.Context, id int32) error {
		restoredIDs = append(restoredIDs, id)
		return nil
	}

	rr, req := setupCommentTest(t, 44, nil, queries)
	if err, ok := restoreCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.AdminRestoreCommentCalls) != 1 || queries.AdminRestoreCommentCalls[0].Idcomments != 44 || queries.AdminRestoreCommentCalls[0].Text.String != "body" {
		t.Fatalf("unexpected restore args: %#v", queries.AdminRestoreCommentCalls)
	}

	if len(restoredIDs) != 1 || restoredIDs[0] != 44 {
		t.Fatalf("unexpected restored ids: %#v", restoredIDs)
	}
}
