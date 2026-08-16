package forum

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestLoadForksForViewerOrdersUnreadFirstAndCapsEachComment(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	baseTime := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
	rows := []*db.GetReplyThreadsForListerRow{
		forkRow(1, 10, false, baseTime.Add(5*time.Minute)),
		forkRow(2, 10, true, baseTime.Add(2*time.Minute)),
		forkRow(3, 10, false, baseTime.Add(9*time.Minute)),
		forkRow(4, 10, true, baseTime.Add(8*time.Minute)),
		forkRow(5, 10, true, baseTime.Add(4*time.Minute)),
		forkRow(6, 10, false, baseTime.Add(7*time.Minute)),
		forkRow(7, 10, true, baseTime.Add(6*time.Minute)),
	}
	queries.GetReplyThreadsForListerReturns = rows
	queries.GetCommentsByThreadIdForUserReturns = []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 10, ForumthreadID: 99}}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 42

	groups, total, err := loadForksForViewer(context.Background(), cd, 99, 5)
	if err != nil {
		t.Fatalf("load forks: %v", err)
	}
	if total != len(rows) {
		t.Fatalf("total = %d, want %d", total, len(rows))
	}
	group := groups[10]
	if group == nil || group.Total != len(rows) || len(group.Threads) != 5 {
		t.Fatalf("group = %#v, want total %d and five previews", group, len(rows))
	}
	want := []int32{4, 7, 5, 2, 3}
	for i, threadID := range want {
		if group.Threads[i].ThreadID != threadID {
			t.Errorf("preview[%d] = thread %d, want %d", i, group.Threads[i].ThreadID, threadID)
		}
	}
	if len(queries.GetReplyThreadsForListerCalls) != 1 {
		t.Fatalf("fork queries = %d, want 1", len(queries.GetReplyThreadsForListerCalls))
	}
	if len(queries.GetCommentsByThreadIdForUserCalls) != 1 {
		t.Fatalf("source comment visibility queries = %d, want 1", len(queries.GetCommentsByThreadIdForUserCalls))
	}
	params := queries.GetReplyThreadsForListerCalls[0]
	if params.ViewerID != cd.UserID || params.ViewerMatchID.Int32 != cd.UserID {
		t.Errorf("viewer parameters = %#v, want user %d", params, cd.UserID)
	}
}

func TestLoadForksForViewerDoesNotExposeHiddenSourceCommentID(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetReplyThreadsForListerReturns = []*db.GetReplyThreadsForListerRow{
		forkRow(1, 10, false, time.Now()),
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	groups, total, err := loadForksForViewer(context.Background(), cd, 99, 5)
	if err != nil {
		t.Fatalf("load forks: %v", err)
	}
	if total != 1 || groups[10] != nil || groups[0] == nil || groups[0].Total != 1 {
		t.Fatalf("hidden source comment relationship was exposed: total=%d groups=%#v", total, groups)
	}
}

func TestLoadForksForViewerReturnsQueryError(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetReplyThreadsForListerErr = errors.New("database unavailable")
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	if _, _, err := loadForksForViewer(context.Background(), cd, 9, 5); err == nil {
		t.Fatal("load forks returned nil error")
	}
}

func TestAuthorizedSourceReferenceDoesNotDiscloseHiddenSource(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetThreadLastPosterAndPermsForUserErr = sql.ErrNoRows
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	child := &db.GetThreadLastPosterAndPermsForUserRow{
		ReplyToThreadID:  sql.NullInt32{Int32: 55, Valid: true},
		ReplyToCommentID: sql.NullInt32{Int32: 66, Valid: true},
	}
	if got := authorizedSourceReference(cd, child); got != nil {
		t.Fatalf("hidden source disclosed as %#v", got)
	}
}

func TestAuthorizedSourceReferenceUsesStableCommentAnchorID(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetThreadLastPosterAndPermsForUserReturns = &db.GetThreadLastPosterAndPermsForUserRow{
		Idforumthread:          55,
		ForumtopicIdforumtopic: 5,
	}
	queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments:    66,
		ForumthreadID: 55,
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	child := &db.GetThreadLastPosterAndPermsForUserRow{
		ReplyToThreadID:  sql.NullInt32{Int32: 55, Valid: true},
		ReplyToCommentID: sql.NullInt32{Int32: 66, Valid: true},
	}
	got := authorizedSourceReference(cd, child)
	if got == nil || got.ThreadID != 55 || got.TopicID != 5 || got.CommentID != 66 {
		t.Fatalf("source reference = %#v", got)
	}
}

func TestAuthorizedSourceReferenceFallsBackToVisibleThread(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetThreadLastPosterAndPermsForUserReturns = &db.GetThreadLastPosterAndPermsForUserRow{
		Idforumthread:          55,
		ForumtopicIdforumtopic: 5,
	}
	queries.GetCommentByIdForUserErr = sql.ErrNoRows
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	child := &db.GetThreadLastPosterAndPermsForUserRow{
		ReplyToThreadID:  sql.NullInt32{Int32: 55, Valid: true},
		ReplyToCommentID: sql.NullInt32{Int32: 66, Valid: true},
	}
	got := authorizedSourceReference(cd, child)
	if got == nil || got.ThreadID != 55 || got.CommentID != 0 {
		t.Fatalf("thread-only source reference = %#v", got)
	}
}

func forkRow(threadID, commentID int32, unread bool, lastAddition time.Time) *db.GetReplyThreadsForListerRow {
	isUnread := int32(0)
	if unread {
		isUnread = 1
	}
	return &db.GetReplyThreadsForListerRow{
		Idforumthread:          threadID,
		ForumtopicIdforumtopic: 5,
		ReplyToCommentID:       sql.NullInt32{Int32: commentID, Valid: true},
		Lastaddition:           sql.NullTime{Time: lastAddition, Valid: true},
		IsUnread:               isUnread,
	}
}
