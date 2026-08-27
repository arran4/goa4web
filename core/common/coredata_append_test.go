package common

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestCanAppendToComment(t *testing.T) {
	tests := []struct {
		name           string
		isPrivate      bool
		cmt            *db.GetCommentsByThreadIdForUserRow
		config         *config.RuntimeConfig
		setupStub      func(stub *db.QuerierStub)
		want           bool
		hasGrant       bool
		threadID       int32
		expectedItem   string
		expectedItemID int32
	}{
		{
			name:      "public eligible -> true",
			isPrivate: false,
			cmt: &db.GetCommentsByThreadIdForUserRow{
				Idcomments:    10,
				ForumthreadID: 100,
				IsOwner:       true,
				Written:       sql.NullTime{Time: time.Now().Add(-10 * time.Minute), Valid: true},
			},
			config:         &config.RuntimeConfig{ForumPostAppendWindow: 60},
			hasGrant:       true,
			expectedItem:   consts.PermissionItemThread.String(),
			expectedItemID: 100,
			want:           true,
			setupStub: func(stub *db.QuerierStub) {
				stub.HasOtherUserReadItemAtOrBeyondFn = func(ctx context.Context, arg db.HasOtherUserReadItemAtOrBeyondParams) (bool, error) {
					if arg.Item != consts.PermissionItemThread.String() || arg.ItemID != 100 || arg.LastCommentID != 10 {
						t.Errorf("Unexpected read marker check params: %v", arg)
					}
					return false, nil
				}
				stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
					return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 10}}, nil
				}
			},
		},
		{
			name:      "private eligible -> true",
			isPrivate: true,
			cmt: &db.GetCommentsByThreadIdForUserRow{
				Idcomments:    11,
				ForumthreadID: 101,
				IsOwner:       true,
				Written:       sql.NullTime{Time: time.Now().Add(-10 * time.Minute), Valid: true},
			},
			config:         &config.RuntimeConfig{PrivateForumPostAppendWindow: 60},
			hasGrant:       true,
			expectedItem:   consts.PermissionItemThread.String(),
			expectedItemID: 101,
			want:           true,
			setupStub: func(stub *db.QuerierStub) {
				stub.HasOtherUserReadItemAtOrBeyondFn = func(ctx context.Context, arg db.HasOtherUserReadItemAtOrBeyondParams) (bool, error) {
					if arg.Item != consts.PermissionItemThread.String() || arg.ItemID != 101 || arg.LastCommentID != 11 {
						t.Errorf("Unexpected read marker check params: %v", arg)
					}
					return false, nil
				}
				stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
					return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 11}}, nil
				}
			},
		},
		{
			name:      "no append grant -> false",
			cmt:       &db.GetCommentsByThreadIdForUserRow{IsOwner: true},
			config:    &config.RuntimeConfig{ForumPostAppendWindow: 60},
			hasGrant:  false,
			want:      false,
			setupStub: func(stub *db.QuerierStub) {},
		},
		{
			name:      "disabled/zero window -> false",
			cmt:       &db.GetCommentsByThreadIdForUserRow{IsOwner: true},
			config:    &config.RuntimeConfig{ForumPostAppendWindow: 0},
			hasGrant:  true,
			want:      false,
			setupStub: func(stub *db.QuerierStub) {},
		},
		{
			name: "outside window -> false",
			cmt: &db.GetCommentsByThreadIdForUserRow{
				IsOwner: true,
				Written: sql.NullTime{Time: time.Now().Add(-120 * time.Minute), Valid: true},
			},
			config:    &config.RuntimeConfig{ForumPostAppendWindow: 60},
			hasGrant:  true,
			want:      false,
			setupStub: func(stub *db.QuerierStub) {},
		},
		{
			name:      "not owner -> false",
			cmt:       &db.GetCommentsByThreadIdForUserRow{IsOwner: false},
			config:    &config.RuntimeConfig{ForumPostAppendWindow: 60},
			want:      false,
			setupStub: func(stub *db.QuerierStub) {},
		},
		{
			name: "no longer final -> false",
			cmt: &db.GetCommentsByThreadIdForUserRow{
				Idcomments:    10,
				ForumthreadID: 100,
				IsOwner:       true,
				Written:       sql.NullTime{Time: time.Now(), Valid: true},
			},
			config:   &config.RuntimeConfig{ForumPostAppendWindow: 60},
			hasGrant: true,
			want:     false,
			setupStub: func(stub *db.QuerierStub) {
				stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
					// 10 is not final, 11 is
					return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 10}, {Idcomments: 11}}, nil
				}
			},
		},
		{
			name: "other marker at candidate -> false",
			cmt: &db.GetCommentsByThreadIdForUserRow{
				Idcomments:    10,
				ForumthreadID: 100,
				IsOwner:       true,
				Written:       sql.NullTime{Time: time.Now(), Valid: true},
			},
			config:   &config.RuntimeConfig{ForumPostAppendWindow: 60},
			hasGrant: true,
			want:     false,
			setupStub: func(stub *db.QuerierStub) {
				stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
					return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 10}}, nil
				}
				stub.HasOtherUserReadItemAtOrBeyondFn = func(ctx context.Context, arg db.HasOtherUserReadItemAtOrBeyondParams) (bool, error) {
					return true, nil // true means someone else read it
				}
			},
		},
		{
			name: "ThreadComments query error -> false",
			cmt: &db.GetCommentsByThreadIdForUserRow{
				Idcomments:    10,
				ForumthreadID: 100,
				IsOwner:       true,
				Written:       sql.NullTime{Time: time.Now(), Valid: true},
			},
			config:   &config.RuntimeConfig{ForumPostAppendWindow: 60},
			hasGrant: true,
			want:     false,
			setupStub: func(stub *db.QuerierStub) {
				stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
					return nil, sql.ErrConnDone
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &db.QuerierStub{}
			tc.setupStub(stub)

			cd := &CoreData{
				UserID:  1,
				Config:  tc.config,
				ctx:     context.Background(),
				queries: stub,
			}

			// Mock CurrentTopic to simulate private forum
			if tc.isPrivate {
				cd.currentTopicID = 50
				stub.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
					if arg.Idforumtopic == 50 {
						return &db.GetForumTopicByIdForUserRow{Handler: "private"}, nil
					}
					return nil, sql.ErrNoRows
				}
			} else {
				cd.currentTopicID = 50
				stub.GetForumTopicByIdFn = func(ctx context.Context, idforumtopic int32) (*db.Forumtopic, error) {
					if idforumtopic == 50 {
						return &db.Forumtopic{Handler: "forum"}, nil
					}
					return nil, sql.ErrNoRows
				}
			}

			// Mock HasGrant manually by hijacking the method or mocking SystemCheckGrant?
			// Actually HasGrant delegates to SystemCheckGrant
			stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
				if arg.Action == "append" {
					if tc.hasGrant {
						return 1, nil
					}
					return 0, nil
				}
				return 0, nil
			}

			got := cd.CanAppendToComment(tc.cmt)
			if got != tc.want {
				t.Errorf("CanAppendToComment() = %v, want %v", got, tc.want)
			}
		})
	}
}
