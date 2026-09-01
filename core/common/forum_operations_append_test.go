package common

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

const (
	// forumReplyAppendActorID identifies the test reply author.
	forumReplyAppendActorID int32 = 7
	// forumReplyAppendTopicID identifies the test forum topic.
	forumReplyAppendTopicID int32 = 37
	// forumReplyAppendThreadID identifies the test forum thread.
	forumReplyAppendThreadID int32 = 412
	// forumReplyAppendCommentID identifies the append candidate.
	forumReplyAppendCommentID int32 = 200
)

func newForumReplyAppendCoreData(t *testing.T, cfg *config.RuntimeConfig, private bool) (*CoreData, *db.QuerierStub, *eventbus.TaskEvent) {
	t.Helper()
	q := testhelpers.NewQuerierStub()
	handler := "forum"
	if private {
		handler = "private"
	}
	q.GetThreadLastPosterAndPermsForUserFn = func(context.Context, db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          forumReplyAppendThreadID,
			ForumtopicIdforumtopic: forumReplyAppendTopicID,
			Firstpost:              100,
		}, nil
	}
	q.GetForumTopicByIdForUserFn = func(context.Context, db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: forumReplyAppendTopicID,
			Handler:      handler,
			Title:        sql.NullString{String: "Append topic", Valid: true},
		}, nil
	}
	q.GetThreadBySectionThreadIDForReplierFn = func(context.Context, db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
		return &db.Forumthread{Idforumthread: forumReplyAppendThreadID}, nil
	}
	q.GetCommentsByThreadIdForUserFn = func(context.Context, db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{{
			Idcomments:    forumReplyAppendCommentID,
			ForumthreadID: forumReplyAppendThreadID,
			UsersIdusers:  forumReplyAppendActorID,
			IsOwner:       true,
			Written:       sql.NullTime{Time: time.Now().Add(-time.Minute), Valid: true},
		}}, nil
	}
	q.GetCommentByIdForUserFn = func(_ context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		if arg.ID == 100 {
			return &db.GetCommentByIdForUserRow{Idcomments: 100, Text: sql.NullString{String: "Opening post", Valid: true}}, nil
		}
		return nil, sql.ErrNoRows
	}
	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := NewCoreData(context.Background(), q, cfg, WithEvent(evt), WithUserRoles([]string{"user"}))
	cd.UserID = forumReplyAppendActorID
	return cd, q, evt
}

func TestReplyForumThreadSuccessfulAppendUsesCanonicalTextAndImages(t *testing.T) {
	cfg := &config.RuntimeConfig{ForumPostAppendWindow: 60}
	cd, q, evt := newForumReplyAppendCoreData(t, cfg, false)
	imageID := "abcd1234.jpg"
	imagePath := path.Join("/", imageID[:2], imageID[2:4], imageID)
	submitted := "new segment [img image:" + imageID + "]"
	canonical := "old segment\n\n[hr]\n\n" + submitted
	q.ListUploadedImagePathsByUserFn = func(context.Context, db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
		return []sql.NullString{{String: imagePath, Valid: true}}, nil
	}
	q.AppendCommentInSectionForCommenterFn = func(_ context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		if arg.Section != "forum" || arg.ItemType.String != "topic" || arg.ItemID.Int32 != forumReplyAppendTopicID {
			t.Fatalf("append grant target = %#v", arg)
		}
		if arg.CommentID != forumReplyAppendCommentID || arg.ForumthreadID != forumReplyAppendThreadID || arg.CommenterID != forumReplyAppendActorID {
			t.Fatalf("append target = %#v", arg)
		}
		return 1, nil
	}
	q.GetCommentByIdForUserFn = func(_ context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		switch arg.ID {
		case forumReplyAppendCommentID:
			return &db.GetCommentByIdForUserRow{Idcomments: arg.ID, Text: sql.NullString{String: canonical, Valid: true}}, nil
		case 100:
			return &db.GetCommentByIdForUserRow{Idcomments: 100, Text: sql.NullString{String: "Opening post", Valid: true}}, nil
		default:
			return nil, sql.ErrNoRows
		}
	}
	q.CreateCommentInSectionForCommenterFn = func(context.Context, db.CreateCommentInSectionForCommenterParams) (int64, error) {
		t.Fatal("normal reply created after successful append")
		return 0, nil
	}

	result, err := cd.ReplyForumThread(context.Background(), ReplyForumThreadParams{
		ActorID: forumReplyAppendActorID, ThreadID: forumReplyAppendThreadID, LanguageID: 1, Text: submitted,
	})
	if err != nil {
		t.Fatalf("ReplyForumThread: %v", err)
	}
	if !result.Appended || result.CommentID != forumReplyAppendCommentID || result.URL != "/forum/topic/37/thread/412#c1" {
		t.Fatalf("result = %#v", result)
	}
	if len(q.CreateThreadImageCalls) != 1 || q.CreateThreadImageCalls[0].Path.String != imagePath {
		t.Fatalf("thread image calls = %#v", q.CreateThreadImageCalls)
	}
	index, ok := evt.Data[searchworker.EventKey].(searchworker.IndexEventData)
	if !ok || index.ID != forumReplyAppendCommentID || index.Text != canonical {
		t.Fatalf("search event = %#v", evt.Data[searchworker.EventKey])
	}
	count, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData)
	if !ok || count.CommentID != forumReplyAppendCommentID || count.ThreadID != forumReplyAppendThreadID || count.TopicID != forumReplyAppendTopicID {
		t.Fatalf("metadata event = %#v", evt.Data[postcountworker.EventKey])
	}
	if got := evt.Data["Body"]; got != canonical {
		t.Fatalf("notification body = %#v", got)
	}
}

func TestReplyForumThreadAppendZeroRowsFallsBackToCreate(t *testing.T) {
	cd, q, evt := newForumReplyAppendCoreData(t, &config.RuntimeConfig{ForumPostAppendWindow: 60}, false)
	q.AppendCommentInSectionForCommenterFn = func(context.Context, db.AppendCommentInSectionForCommenterParams) (int64, error) {
		return 0, nil
	}
	q.CreateCommentInSectionForCommenterFn = func(context.Context, db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return 300, nil
	}

	result, err := cd.ReplyForumThread(context.Background(), ReplyForumThreadParams{
		ActorID: forumReplyAppendActorID, ThreadID: forumReplyAppendThreadID, Text: "new reply",
	})
	if err != nil {
		t.Fatalf("ReplyForumThread: %v", err)
	}
	if result.Appended || result.CommentID != 300 || result.URL != "/forum/topic/37/thread/412#c2" {
		t.Fatalf("result = %#v", result)
	}
	index, ok := evt.Data[searchworker.EventKey].(searchworker.IndexEventData)
	if !ok || index.ID != 300 || index.Text != "new reply" {
		t.Fatalf("search event = %#v", evt.Data[searchworker.EventKey])
	}
}

func TestReplyForumThreadAppendErrorNeverCreatesReply(t *testing.T) {
	cd, q, _ := newForumReplyAppendCoreData(t, &config.RuntimeConfig{ForumPostAppendWindow: 60}, false)
	wantErr := errors.New("append database unavailable")
	q.AppendCommentInSectionForCommenterFn = func(context.Context, db.AppendCommentInSectionForCommenterParams) (int64, error) {
		return 0, wantErr
	}
	created := false
	q.CreateCommentInSectionForCommenterFn = func(context.Context, db.CreateCommentInSectionForCommenterParams) (int64, error) {
		created = true
		return 300, nil
	}

	_, err := cd.ReplyForumThread(context.Background(), ReplyForumThreadParams{
		ActorID: forumReplyAppendActorID, ThreadID: forumReplyAppendThreadID, Text: "new reply",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped %v", err, wantErr)
	}
	if created {
		t.Fatal("normal reply created after append database error")
	}
}

func TestReplyForumThreadAppendReloadFailureIsIrreversible(t *testing.T) {
	cd, q, evt := newForumReplyAppendCoreData(t, &config.RuntimeConfig{ForumPostAppendWindow: 60}, false)
	q.AppendCommentInSectionForCommenterFn = func(context.Context, db.AppendCommentInSectionForCommenterParams) (int64, error) {
		return 1, nil
	}
	q.GetCommentByIdForUserFn = func(_ context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		if arg.ID == forumReplyAppendCommentID {
			return nil, errors.New("reload failed")
		}
		return &db.GetCommentByIdForUserRow{Idcomments: 100, Text: sql.NullString{String: "Opening post", Valid: true}}, nil
	}
	created := false
	q.CreateCommentInSectionForCommenterFn = func(context.Context, db.CreateCommentInSectionForCommenterParams) (int64, error) {
		created = true
		return 300, nil
	}

	result, err := cd.ReplyForumThread(context.Background(), ReplyForumThreadParams{
		ActorID: forumReplyAppendActorID, ThreadID: forumReplyAppendThreadID, Text: "submitted fragment",
	})
	if err != nil {
		t.Fatalf("ReplyForumThread: %v", err)
	}
	if !result.Appended || result.CommentID != forumReplyAppendCommentID || created {
		t.Fatalf("result = %#v, created = %v", result, created)
	}
	if _, ok := evt.Data[searchworker.EventKey]; ok {
		t.Fatalf("reload failure indexed guessed text: %#v", evt.Data[searchworker.EventKey])
	}
	if _, ok := evt.Data[postcountworker.EventKey]; !ok {
		t.Fatal("reload failure omitted safe metadata refresh")
	}
	if got := evt.Data["Body"]; got != "" {
		t.Fatalf("reload failure notification body = %#v", got)
	}
}

func TestReplyForumThreadAppendWindowsAndScopesAreIndependent(t *testing.T) {
	tests := []struct {
		name             string
		private          bool
		publicWindow     int
		privateWindow    int
		wantAppend       bool
		wantSection      string
		wantItem         string
		wantItemID       int32
		wantCreateAction string
	}{
		{name: "public enabled private disabled", publicWindow: 60, wantAppend: true, wantSection: "forum", wantItem: "topic", wantItemID: forumReplyAppendTopicID},
		{name: "public disabled private enabled", private: true, privateWindow: 60, wantAppend: true, wantSection: "privateforum_thread", wantItem: "thread", wantItemID: forumReplyAppendThreadID},
		{name: "public zero only disables public", publicWindow: 0, privateWindow: 60, wantCreateAction: "reply"},
		{name: "private zero only disables private", private: true, publicWindow: 60, privateWindow: 0, wantCreateAction: "reply"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.RuntimeConfig{ForumPostAppendWindow: tc.publicWindow, PrivateForumPostAppendWindow: tc.privateWindow}
			cd, q, _ := newForumReplyAppendCoreData(t, cfg, tc.private)
			appendCalled := false
			q.AppendCommentInSectionForCommenterFn = func(_ context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
				appendCalled = true
				if arg.Section != tc.wantSection || arg.ItemType.String != tc.wantItem || arg.ItemID.Int32 != tc.wantItemID {
					t.Fatalf("append scope = %#v", arg)
				}
				return 0, nil
			}
			q.CreateCommentInSectionForCommenterFn = func(_ context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
				if tc.wantCreateAction != "" && arg.Action != tc.wantCreateAction {
					t.Fatalf("create action = %q", arg.Action)
				}
				return 300, nil
			}
			_, err := cd.ReplyForumThread(context.Background(), ReplyForumThreadParams{
				ActorID: forumReplyAppendActorID, ThreadID: forumReplyAppendThreadID, Text: "reply", Private: tc.private, EnforceHandler: true,
			})
			if err != nil {
				t.Fatalf("ReplyForumThread: %v", err)
			}
			if appendCalled != tc.wantAppend {
				t.Fatalf("append called = %v, want %v", appendCalled, tc.wantAppend)
			}
		})
	}
}

func TestCanAppendToCommentEligibilityAndMarkerScope(t *testing.T) {
	tests := []struct {
		name          string
		private       bool
		window        int
		owner         bool
		written       time.Time
		grantErr      error
		comments      []*db.GetCommentsByThreadIdForUserRow
		commentsErr   error
		hasOtherRead  bool
		markerErr     error
		want          bool
		wantSection   string
		wantGrantItem string
		wantGrantID   int32
	}{
		{name: "public eligible", window: 60, owner: true, written: time.Now(), comments: []*db.GetCommentsByThreadIdForUserRow{{Idcomments: forumReplyAppendCommentID}}, want: true, wantSection: "forum", wantGrantItem: "topic", wantGrantID: forumReplyAppendTopicID},
		{name: "private eligible", private: true, window: 60, owner: true, written: time.Now(), comments: []*db.GetCommentsByThreadIdForUserRow{{Idcomments: forumReplyAppendCommentID}}, want: true, wantSection: "privateforum_thread", wantGrantItem: "thread", wantGrantID: forumReplyAppendThreadID},
		{name: "zero window", owner: true, written: time.Now()},
		{name: "not owner", window: 60, written: time.Now()},
		{name: "outside window", window: 60, owner: true, written: time.Now().Add(-2 * time.Hour)},
		{name: "missing append grant", window: 60, owner: true, written: time.Now(), grantErr: sql.ErrNoRows},
		{name: "not final", window: 60, owner: true, written: time.Now(), comments: []*db.GetCommentsByThreadIdForUserRow{{Idcomments: forumReplyAppendCommentID}, {Idcomments: forumReplyAppendCommentID + 1}}},
		{name: "comment list error", window: 60, owner: true, written: time.Now(), commentsErr: errors.New("comments failed")},
		{name: "other user read candidate", window: 60, owner: true, written: time.Now(), comments: []*db.GetCommentsByThreadIdForUserRow{{Idcomments: forumReplyAppendCommentID}}, hasOtherRead: true},
		{name: "marker error fails closed", window: 60, owner: true, written: time.Now(), comments: []*db.GetCommentsByThreadIdForUserRow{{Idcomments: forumReplyAppendCommentID}}, markerErr: errors.New("markers failed")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.RuntimeConfig{}
			if tc.private {
				cfg.PrivateForumPostAppendWindow = tc.window
			} else {
				cfg.ForumPostAppendWindow = tc.window
			}
			cd, q, _ := newForumReplyAppendCoreData(t, cfg, tc.private)
			cd.SetCurrentThreadAndTopic(forumReplyAppendThreadID, forumReplyAppendTopicID)
			q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
				if tc.wantSection != "" && (arg.Section != tc.wantSection || arg.Item.String != tc.wantGrantItem || arg.ItemID.Int32 != tc.wantGrantID || arg.Action != "append") {
					t.Fatalf("grant target = %#v", arg)
				}
				if tc.grantErr != nil {
					return 0, tc.grantErr
				}
				return 1, nil
			}
			q.GetCommentsByThreadIdForUserFn = func(context.Context, db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
				return tc.comments, tc.commentsErr
			}
			q.SystemHasOtherUserReadItemAtOrBeyondFn = func(_ context.Context, arg db.SystemHasOtherUserReadItemAtOrBeyondParams) (bool, error) {
				if arg.Item != "thread" || arg.ItemID != forumReplyAppendThreadID || arg.UserID != forumReplyAppendActorID || arg.LastCommentID != forumReplyAppendCommentID {
					t.Fatalf("marker target = %#v", arg)
				}
				return tc.hasOtherRead, tc.markerErr
			}
			candidate := &db.GetCommentsByThreadIdForUserRow{
				Idcomments: forumReplyAppendCommentID, ForumthreadID: forumReplyAppendThreadID,
				UsersIdusers: forumReplyAppendActorID, IsOwner: tc.owner,
				Written: sql.NullTime{Time: tc.written, Valid: !tc.written.IsZero()},
			}
			if got := cd.CanAppendToComment(candidate); got != tc.want {
				t.Fatalf("CanAppendToComment() = %v, want %v", got, tc.want)
			}
		})
	}
}
