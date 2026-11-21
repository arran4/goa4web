package common

import (
	"context"
	"database/sql"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

type recordingForumReplyNotifier struct {
	intents []ForumReplyNotificationIntent
}

func (r *recordingForumReplyNotifier) NotifyForumReply(_ context.Context, intent ForumReplyNotificationIntent) error {
	r.intents = append(r.intents, intent)
	return nil
}

func TestCoreCommonDoesNotImportInternalNotifications(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read core/common directory: %v", err)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == "github.com/arran4/goa4web/internal/notifications" || strings.HasPrefix(path, "github.com/arran4/goa4web/internal/notifications/") {
				t.Fatalf("%s imports %q, which would create a core/common <-> internal/notifications dependency cycle", entry.Name(), path)
			}
		}
	}
}

func TestForumReplyService_HeadlessCanonicalExecutionAndAppendActivity(t *testing.T) {
	const (
		actorID   int32 = 11
		topicID   int32 = 22
		threadID  int32 = 33
		commentID int32 = 444
	)

	currentText := "First reply body"
	appendMode := false

	qs := &db.QuerierStub{
		GetThreadLastPosterAndPermsForUserFn: func(_ context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
			if arg.ThreadID != threadID {
				return nil, sql.ErrNoRows
			}
			return &db.GetThreadLastPosterAndPermsForUserRow{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
				Firstpost:              1,
				Comments:               sql.NullInt32{Int32: 2, Valid: true},
			}, nil
		},
		GetForumTopicByIdForUserFn: func(_ context.Context, _ db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: topicID,
				Title:        sql.NullString{String: "Architecture Topic", Valid: true},
				Handler:      "private",
			}, nil
		},
		GetThreadBySectionThreadIDForReplierFn: func(_ context.Context, _ db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
			return &db.Forumthread{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
			}, nil
		},
		GetCommentsByThreadIdForUserFn: func(_ context.Context, _ db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
			if !appendMode {
				return nil, nil
			}
			return []*db.GetCommentsByThreadIdForUserRow{
				{Idcomments: 1, Text: sql.NullString{String: "Thread opener text", Valid: true}},
				{Idcomments: commentID, Text: sql.NullString{String: currentText, Valid: true}},
			}, nil
		},
		CreateCommentInSectionForCommenterFn: func(_ context.Context, _ db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return int64(commentID), nil
		},
		AppendCommentInSectionForCommenterFn: func(_ context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
			if !appendMode {
				return 0, nil
			}
			currentText = currentText + "\n\n[hr]\n\n" + fmt.Sprint(arg.Text)
			return 1, nil
		},
		GetCommentByIdForUserFn: func(_ context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
			if arg.ID == 1 {
				return &db.GetCommentByIdForUserRow{
					Idcomments: 1,
					Text:       sql.NullString{String: "Thread opener text", Valid: true},
				}, nil
			}
			return &db.GetCommentByIdForUserRow{
				Idcomments: commentID,
				Text:       sql.NullString{String: currentText, Valid: true},
			}, nil
		},
		SystemGetUserByIDFn: func(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  id,
				Username: sql.NullString{String: "bob", Valid: true},
			}, nil
		},
		UpsertContentReadMarkerFn:                  func(_ context.Context, _ db.UpsertContentReadMarkerParams) error { return nil },
		ClearUnreadContentPrivateLabelExceptUserFn: func(_ context.Context, _ db.ClearUnreadContentPrivateLabelExceptUserParams) error { return nil },
	}

	cfg := config.NewRuntimeConfig()
	cfg.BaseURL = "https://forum.example.test"
	cfg.PrivateForumPostAppendWindow = 60
	cd := NewCoreData(context.Background(), qs, cfg)
	notifier := &recordingForumReplyNotifier{}
	svc := NewForumReplyService(cd, notifier)

	// 1. Mismatched TopicID must fail closed with ForumResourceNotFoundError.
	if _, err := svc.Reply(context.Background(), ForumReplyCommand{
		ActorID:                actorID,
		TopicID:                topicID + 999,
		ThreadID:               threadID,
		LanguageID:             1,
		Text:                   "Should fail",
		Private:                true,
		EnforceHandler:         true,
		SynchronousSideEffects: true,
	}); err == nil {
		t.Fatal("expected ForumResourceNotFoundError for mismatched TopicID, got nil")
	}

	// 2. First headless reply creates a comment and emits canonical intent.
	res1, err := svc.Reply(context.Background(), ForumReplyCommand{
		ActorID:        actorID,
		TopicID:        topicID,
		ThreadID:       threadID,
		LanguageID:     1,
		Text:           "First reply body",
		At:             time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
		Private:        true,
		EnforceHandler: true,
		BasePath:       "/private/api",
	})
	if err != nil {
		t.Fatalf("first Reply error: %v", err)
	}
	if res1.Appended {
		t.Fatal("expected first reply not to be appended")
	}
	if res1.ActivityID != "comment:444" {
		t.Fatalf("first reply ActivityID = %q, want comment:444", res1.ActivityID)
	}
	if res1.CanonicalPath != "/private/topic/22/thread/33" {
		t.Fatalf("first reply CanonicalPath = %q, want /private/topic/22/thread/33", res1.CanonicalPath)
	}
	if res1.CommentURL != "/private/topic/22/thread/33#c444" {
		t.Fatalf("first reply CommentURL = %q, want /private/topic/22/thread/33#c444", res1.CommentURL)
	}

	// 3. Consecutive reply by same author appends to comment 444 with a distinct ActivityID.
	appendMode = true
	res2, err := svc.Reply(context.Background(), ForumReplyCommand{
		ActorID:        actorID,
		TopicID:        topicID,
		ThreadID:       threadID,
		LanguageID:     1,
		Text:           "Appended follow-up body",
		At:             time.Date(2026, 8, 1, 10, 5, 0, 0, time.UTC),
		Private:        true,
		EnforceHandler: true,
		BasePath:       "/private/api",
	})
	if err != nil {
		t.Fatalf("second (appended) Reply error: %v", err)
	}
	if !res2.Appended {
		t.Fatal("expected second reply to be appended")
	}
	if res2.CommentID != commentID {
		t.Fatalf("second reply CommentID = %d, want %d", res2.CommentID, commentID)
	}
	if res2.ActivityID == res1.ActivityID || res2.ActivityID != "comment:444:append:1" {
		t.Fatalf("expected distinct append ActivityID comment:444:append:1, got %q (first=%q)", res2.ActivityID, res1.ActivityID)
	}
	if res2.AppendID != 1 {
		t.Fatalf("expected AppendID 1, got %d", res2.AppendID)
	}
	if len(notifier.intents) != 2 {
		t.Fatalf("notifier intents = %d, want 2", len(notifier.intents))
	}
	if notifier.intents[0].ActivityID != res1.ActivityID || notifier.intents[1].ActivityID != res2.ActivityID {
		t.Fatalf("recorded intents = %+v, want %q then %q", notifier.intents, res1.ActivityID, res2.ActivityID)
	}
}
