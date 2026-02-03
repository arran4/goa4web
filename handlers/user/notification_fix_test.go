package user

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestFixNotificationLinkAndGetData(t *testing.T) {
	qs := &db.QuerierStub{}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	// Mock data
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		if arg.Idforumtopic == 15 {
			return &db.GetForumTopicByIdForUserRow{
				Handler: "private",
			}, nil
		}
		return nil, sql.ErrNoRows
	}

	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		if arg.ThreadID == 47 {
			return &db.GetThreadLastPosterAndPermsRow{
				Firstpost: 100,
			}, nil
		}
		return nil, sql.ErrNoRows
	}

	qs.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		if arg.ID == 100 {
			return &db.GetCommentByIdForUserRow{
				Text: sql.NullString{String: "Thread Title Text", Valid: true},
			}, nil
		}
		return nil, sql.ErrNoRows
	}

	tests := []struct {
		name             string
		link             string
		wantFixedLink    string
		wantThreadTitle  string
		wantSectionTitle string
	}{
		{
			name:             "Broken Private Link",
			link:             "/private/topic/15/thread/47/reply",
			wantFixedLink:    "/private/topic/15/thread/47#bottom",
			wantThreadTitle:  "Thread Title Text",
			wantSectionTitle: "Private Forum",
		},
		{
			name:             "Broken Public Link",
			link:             "/topic/15/thread/47/reply",
			wantFixedLink:    "/topic/15/thread/47#bottom",
			wantThreadTitle:  "Thread Title Text",
			wantSectionTitle: "Private Forum", // Topic 15 is mocked as private regardless of link prefix
		},
		{
			name:             "Correct Link",
			link:             "/private/topic/15/thread/47#c13",
			wantFixedLink:    "/private/topic/15/thread/47#c13",
			wantThreadTitle:  "Thread Title Text",
			wantSectionTitle: "Private Forum",
		},
		{
			name:             "Unknown Thread",
			link:             "/topic/99/thread/99/reply",
			wantFixedLink:    "/topic/99/thread/99#bottom",
			wantThreadTitle:  "",
			wantSectionTitle: "Forum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFixedLink, gotThreadTitle, gotSectionTitle, err := fixNotificationLinkAndGetData(cd, tt.link)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotFixedLink != tt.wantFixedLink {
				t.Errorf("got fixed link %q, want %q", gotFixedLink, tt.wantFixedLink)
			}
			if gotThreadTitle != tt.wantThreadTitle {
				t.Errorf("got thread title %q, want %q", gotThreadTitle, tt.wantThreadTitle)
			}
			if gotSectionTitle != tt.wantSectionTitle {
				t.Errorf("got section title %q, want %q", gotSectionTitle, tt.wantSectionTitle)
			}
		})
	}
}
