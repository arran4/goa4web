package admin

import (
	"context"
	"database/sql"
	"fmt"
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
)

type commentTaskQueries struct {
	db.Querier
	commentID       int32
	comment         *db.GetCommentByIdForUserRow
	deactivated     bool
	deactivatedRows []*db.AdminListDeactivatedCommentsRow
	scrubArgs       []db.AdminScrubCommentParams
	archiveArgs     []db.AdminArchiveCommentParams
	restoreArgs     []db.AdminRestoreCommentParams
	restoredIDs     []int32
}

func (q *commentTaskQueries) GetCommentByIdForUser(_ context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
	if arg.ID != q.commentID {
		return nil, fmt.Errorf("unexpected comment id: %d", arg.ID)
	}
	return q.comment, nil
}

func (q *commentTaskQueries) AdminScrubComment(_ context.Context, arg db.AdminScrubCommentParams) error {
	q.scrubArgs = append(q.scrubArgs, arg)
	return nil
}

func (q *commentTaskQueries) AdminIsCommentDeactivated(_ context.Context, id int32) (bool, error) {
	if id != q.commentID {
		return false, fmt.Errorf("unexpected comment id: %d", id)
	}
	return q.deactivated, nil
}

func (q *commentTaskQueries) AdminArchiveComment(_ context.Context, arg db.AdminArchiveCommentParams) error {
	q.archiveArgs = append(q.archiveArgs, arg)
	return nil
}

func (q *commentTaskQueries) AdminListDeactivatedComments(context.Context, db.AdminListDeactivatedCommentsParams) ([]*db.AdminListDeactivatedCommentsRow, error) {
	return q.deactivatedRows, nil
}

func (q *commentTaskQueries) AdminRestoreComment(_ context.Context, arg db.AdminRestoreCommentParams) error {
	q.restoreArgs = append(q.restoreArgs, arg)
	return nil
}

func (q *commentTaskQueries) AdminMarkCommentRestored(_ context.Context, id int32) error {
	q.restoredIDs = append(q.restoredIDs, id)
	return nil
}

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

func TestDeleteCommentTask_UsesURLParam(t *testing.T) {
	queries := &commentTaskQueries{
		commentID: 15,
		comment: &db.GetCommentByIdForUserRow{
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
		},
	}
	rr, req := setupCommentTest(t, 15, nil, queries)
	if err, ok := deleteCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.scrubArgs) != 1 || queries.scrubArgs[0].Idcomments != 15 {
		t.Fatalf("unexpected scrub args: %#v", queries.scrubArgs)
	}
}

func TestEditCommentTask_UsesURLParam(t *testing.T) {
	body := url.Values{"replytext": {"updated"}}
	queries := &commentTaskQueries{
		commentID: 22,
		comment: &db.GetCommentByIdForUserRow{
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
		},
	}
	rr, req := setupCommentTest(t, 22, body, queries)
	if err, ok := editCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.scrubArgs) != 1 || queries.scrubArgs[0].Idcomments != 22 || queries.scrubArgs[0].Text.String != "updated" {
		t.Fatalf("unexpected scrub args: %#v", queries.scrubArgs)
	}
}

func TestDeactivateCommentTask_UsesURLParam(t *testing.T) {
	queries := &commentTaskQueries{
		commentID:   33,
		deactivated: false,
		comment: &db.GetCommentByIdForUserRow{
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
		},
	}
	rr, req := setupCommentTest(t, 33, nil, queries)
	if err, ok := deactivateCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.archiveArgs) != 1 || queries.archiveArgs[0].Idcomments != 33 {
		t.Fatalf("unexpected archive args: %#v", queries.archiveArgs)
	}
	if len(queries.scrubArgs) != 1 || queries.scrubArgs[0].Idcomments != 33 {
		t.Fatalf("unexpected scrub args: %#v", queries.scrubArgs)
	}
}

func TestRestoreCommentTask_UsesURLParam(t *testing.T) {
	queries := &commentTaskQueries{
		commentID:   44,
		deactivated: true,
		comment: &db.GetCommentByIdForUserRow{
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
		},
		deactivatedRows: []*db.AdminListDeactivatedCommentsRow{{
			Idcomments: 44,
			Text:       sql.NullString{String: "body", Valid: true},
		}},
	}
	rr, req := setupCommentTest(t, 44, nil, queries)
	if err, ok := restoreCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.restoreArgs) != 1 || queries.restoreArgs[0].Idcomments != 44 || queries.restoreArgs[0].Text.String != "body" {
		t.Fatalf("unexpected restore args: %#v", queries.restoreArgs)
	}
	if len(queries.restoredIDs) != 1 || queries.restoredIDs[0] != 44 {
		t.Fatalf("unexpected restored ids: %#v", queries.restoredIDs)
	}
}
