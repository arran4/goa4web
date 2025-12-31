package forum

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestTopicsPage_ThreadLinks(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"

	now := time.Date(2024, time.January, 2, 15, 4, 5, 0, time.UTC)
	queries := &forumTopicPageQuerierFake{
		categories: []*db.Forumcategory{{
			Idforumcategory:              1,
			ForumcategoryIdforumcategory: 0,
			Title:                        sql.NullString{String: "Category", Valid: true},
		}},
		topic: &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 1,
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "Topic", Valid: true},
			Description:                  sql.NullString{String: "topic description", Valid: true},
			Threads:                      sql.NullInt32{Int32: 1, Valid: true},
			Comments:                     sql.NullInt32{Int32: 2, Valid: true},
			Lastaddition:                 sql.NullTime{Time: now, Valid: true},
			Handler:                      "",
			Lastposterusername:           sql.NullString{String: "last poster", Valid: true},
		},
		threads: []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{{
			Idforumthread:          2,
			Firstpost:              10,
			Lastposter:             0,
			ForumtopicIdforumtopic: 1,
			Comments:               sql.NullInt32{Int32: 0, Valid: true},
			Lastaddition:           sql.NullTime{Time: now, Valid: true},
			Locked:                 sql.NullBool{Bool: false, Valid: true},
			Lastposterusername:     sql.NullString{String: "abc", Valid: true},
			Lastposterid:           sql.NullInt32{Int32: 5, Valid: true},
			Firstpostusername:      sql.NullString{String: "abc", Valid: true},
			Firstpostwritten:       sql.NullTime{Time: now, Valid: true},
			Firstposttext:          sql.NullString{String: "first post", Valid: true},
		}},
		publicLabels: map[int32][]string{
			2: {"hot"},
		},
		ownerLabels: map[int32][]string{},
		privateLabels: map[int32][]privateLabelRow{
			2: {
				{label: "new", invert: true},
				{label: "unread", invert: true},
			},
		},
	}

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodGet, "/forum/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	TopicsPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "href=\"/forum/topic/1/thread/2\"") {
		t.Fatalf("expected thread link, got %q", body)
	}
	if strings.Contains(body, "href=\"//forum") {
		t.Fatalf("unexpected double slash in link: %q", body)
	}
	expectedCalls := []string{
		"GetAllForumCategories",
		"GetForumTopicByIdForUser:1",
		"GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText:1",
		"ListContentPublicLabels:thread:2",
		"ListContentLabelStatus:thread:2",
		"ListContentPrivateLabels:thread:2",
		"ListContentPublicLabels:thread:1",
		"ListContentLabelStatus:thread:1",
	}
	if diff := cmp.Diff(expectedCalls, queries.calls); diff != "" {
		t.Fatalf("unexpected query sequence (-want +got):\n%s", diff)
	}
}

type privateLabelRow struct {
	label  string
	invert bool
}

type forumTopicPageQuerierFake struct {
	db.Querier
	categories    []*db.Forumcategory
	topic         *db.GetForumTopicByIdForUserRow
	threads       []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
	publicLabels  map[int32][]string
	ownerLabels   map[int32][]string
	privateLabels map[int32][]privateLabelRow
	calls         []string
}

var _ db.Querier = (*forumTopicPageQuerierFake)(nil)

func (f *forumTopicPageQuerierFake) record(call string) {
	f.calls = append(f.calls, call)
}

func (f *forumTopicPageQuerierFake) GetAllForumCategories(_ context.Context, _ db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
	f.record("GetAllForumCategories")
	return f.categories, nil
}

func (f *forumTopicPageQuerierFake) GetForumTopicByIdForUser(_ context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
	f.record(fmt.Sprintf("GetForumTopicByIdForUser:%d", arg.Idforumtopic))
	return f.topic, nil
}

func (f *forumTopicPageQuerierFake) SystemCheckRoleGrant(_ context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	f.record(fmt.Sprintf("SystemCheckRoleGrant:%s:%s", arg.Name, arg.Action))
	return 0, nil
}

func (f *forumTopicPageQuerierFake) GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(_ context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	f.record(fmt.Sprintf("GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText:%d", arg.TopicID))
	return f.threads, nil
}

func (f *forumTopicPageQuerierFake) ListContentPublicLabels(_ context.Context, arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
	f.record(fmt.Sprintf("ListContentPublicLabels:%s:%d", arg.Item, arg.ItemID))
	labels := f.publicLabels[arg.ItemID]
	rows := make([]*db.ListContentPublicLabelsRow, 0, len(labels))
	for _, l := range labels {
		rows = append(rows, &db.ListContentPublicLabelsRow{
			Item:   arg.Item,
			ItemID: arg.ItemID,
			Label:  l,
		})
	}
	return rows, nil
}

func (f *forumTopicPageQuerierFake) ListContentLabelStatus(_ context.Context, arg db.ListContentLabelStatusParams) ([]*db.ListContentLabelStatusRow, error) {
	f.record(fmt.Sprintf("ListContentLabelStatus:%s:%d", arg.Item, arg.ItemID))
	labels := f.ownerLabels[arg.ItemID]
	rows := make([]*db.ListContentLabelStatusRow, 0, len(labels))
	for _, l := range labels {
		rows = append(rows, &db.ListContentLabelStatusRow{
			Item:   arg.Item,
			ItemID: arg.ItemID,
			Label:  l,
		})
	}
	return rows, nil
}

func (f *forumTopicPageQuerierFake) ListContentPrivateLabels(_ context.Context, arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
	f.record(fmt.Sprintf("ListContentPrivateLabels:%s:%d", arg.Item, arg.ItemID))
	labels := f.privateLabels[arg.ItemID]
	rows := make([]*db.ListContentPrivateLabelsRow, 0, len(labels))
	for _, l := range labels {
		rows = append(rows, &db.ListContentPrivateLabelsRow{
			Item:   arg.Item,
			ItemID: arg.ItemID,
			UserID: arg.UserID,
			Label:  l.label,
			Invert: l.invert,
		})
	}
	return rows, nil
}
