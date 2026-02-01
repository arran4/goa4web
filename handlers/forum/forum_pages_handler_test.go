package forum

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

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForumPageHandlers(t *testing.T) {
	t.Run("admin topics page", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.AdminCountForumTopicsReturns = 1
		queries.AdminListForumTopicsReturns = []*db.Forumtopic{
			{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 1,
				Title:                        sql.NullString{String: "t", Valid: true},
				Description:                  sql.NullString{String: "d", Valid: true},
				Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
				Handler:                      "",
			},
		}
		queries.GetAllForumCategoriesReturns = []*db.Forumcategory{
			{
				Idforumcategory:              1,
				ForumcategoryIdforumcategory: 0,
				Title:                        sql.NullString{String: "cat", Valid: true},
				Description:                  sql.NullString{String: "desc", Valid: true},
			},
		}
		queries.AdminGetTopicGrantsReturns = []*db.AdminGetTopicGrantsRow{
			{Section: "forum", RoleID: sql.NullInt32{}, RoleName: sql.NullString{}, UserID: sql.NullInt32{}, Username: sql.NullString{}},
		}

		origStore := core.Store
		origName := core.SessionName
		core.Store = sessions.NewCookieStore([]byte("test"))
		core.SessionName = "test-session"
		defer func() {
			core.Store = origStore
			core.SessionName = origName
		}()

		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		r := httptest.NewRequest("GET", "/admin/forum/topics", nil)
		sess := testhelpers.Must(core.Store.New(r, core.SessionName))
		ctx := context.WithValue(r.Context(), core.ContextValues("session"), sess)
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		AdminTopicsPage(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("admin topic page", func(t *testing.T) {
		topicID := 4
		queries := testhelpers.NewQuerierStub()
		queries.GetForumTopicByIdReturns = &db.Forumtopic{
			Idforumtopic:                 int32(topicID),
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "t", Valid: true},
			Description:                  sql.NullString{String: "d", Valid: true},
			Threads:                      sql.NullInt32{Int32: 2, Valid: true},
			Comments:                     sql.NullInt32{Int32: 3, Valid: true},
			Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			Handler:                      "",
		}
		queries.AdminListForumTopicGrantsByTopicIDReturns = []*db.AdminListForumTopicGrantsByTopicIDRow{
			{ID: 1, Section: "forum", Action: "see", RoleName: sql.NullString{String: "Anyone", Valid: true}},
		}

		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		r := httptest.NewRequest("GET", "/admin/forum/topics/topic/"+strconv.Itoa(topicID), nil)
		r = mux.SetURLVars(r, map[string]string{"topic": strconv.Itoa(topicID)})
		r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
		w := httptest.NewRecorder()
		AdminTopicPage(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("admin topic edit form page", func(t *testing.T) {
		topicID := 4
		queries := testhelpers.NewQuerierStub()
		queries.GetForumTopicByIdReturns = &db.Forumtopic{
			Idforumtopic:                 int32(topicID),
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "t", Valid: true},
			Description:                  sql.NullString{String: "d", Valid: true},
			Threads:                      sql.NullInt32{Int32: 2, Valid: true},
			Comments:                     sql.NullInt32{Int32: 3, Valid: true},
			Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			Handler:                      "",
		}
		queries.GetAllForumCategoriesReturns = []*db.Forumcategory{
			{
				Idforumcategory:              1,
				ForumcategoryIdforumcategory: 0,
				Title:                        sql.NullString{String: "cat", Valid: true},
				Description:                  sql.NullString{String: "desc", Valid: true},
			},
		}
		queries.AdminListRolesReturns = []*db.Role{
			{ID: 1, Name: "role", CanLogin: true},
		}

		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		r := httptest.NewRequest("GET", "/admin/forum/topics/topic/"+strconv.Itoa(topicID)+"/edit", nil)
		r = mux.SetURLVars(r, map[string]string{"topic": strconv.Itoa(topicID)})
		r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
		w := httptest.NewRecorder()
		AdminTopicEditFormPage(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("admin category pages", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.GetForumCategoryByIdReturns = &db.Forumcategory{
			Idforumcategory:              1,
			ForumcategoryIdforumcategory: 0,
			LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
			Title:                        sql.NullString{String: "cat", Valid: true},
			Description:                  sql.NullString{String: "desc", Valid: true},
		}
		queries.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameReturns = []*db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow{
			{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 1,
				Title:                        sql.NullString{String: "t", Valid: true},
				Description:                  sql.NullString{String: "d", Valid: true},
				Threads:                      sql.NullInt32{Int32: 0, Valid: true},
				Comments:                     sql.NullInt32{Int32: 0, Valid: true},
				Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
				Handler:                      "",
			},
		}
		queries.GetAllForumCategoriesReturns = []*db.Forumcategory{
			{
				Idforumcategory:              1,
				ForumcategoryIdforumcategory: 0,
				LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
				Title:                        sql.NullString{String: "cat", Valid: true},
				Description:                  sql.NullString{String: "desc", Valid: true},
			},
		}
		queries.AdminListRolesReturns = []*db.Role{
			{ID: 1, Name: "user", CanLogin: true},
		}
		queries.ListGrantsReturns = []*db.Grant{
			{
				ID:      1,
				RoleID:  sql.NullInt32{Int32: 1, Valid: true},
				Section: "forum",
				Item:    sql.NullString{String: "category", Valid: true},
				ItemID:  sql.NullInt32{Int32: 1, Valid: true},
				Action:  "see",
				Active:  true,
			},
		}

		t.Run("category overview links", func(t *testing.T) {
			req, rr := setupRequest(t, queries, "/admin/forum/categories/category/1", map[string]string{"category": "1"})

			AdminCategoryPage(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status=%d", rr.Code)
			}
			body := rr.Body.String()
			if !strings.Contains(body, "/admin/forum/categories/category/1/edit") {
				t.Fatalf("missing edit link")
			}
			if !strings.Contains(body, "/admin/forum/categories/category/1/grants") {
				t.Fatalf("missing grants link")
			}
			if !strings.Contains(body, "<a href=\"/admin/forum/topic/1\">1</a>") {
				t.Fatalf("missing topic link")
			}
		})

		t.Run("category edit form", func(t *testing.T) {
			req, rr := setupRequest(t, queries, "/admin/forum/categories/category/1/edit", map[string]string{"category": "1"})

			AdminCategoryEditPage(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status=%d", rr.Code)
			}
		})

		t.Run("category grants", func(t *testing.T) {
			req, rr := setupRequest(t, queries, "/admin/forum/categories/category/1/grants", map[string]string{"category": "1"})

			AdminCategoryGrantsPage(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status=%d", rr.Code)
			}
		})
	})

	t.Run("topics page thread links", func(t *testing.T) {
		origStore := core.Store
		origName := core.SessionName
		core.Store = sessions.NewCookieStore([]byte("test"))
		core.SessionName = "test-session"
		defer func() {
			core.Store = origStore
			core.SessionName = origName
		}()

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
				Firstpostuserid:        sql.NullInt32{Int32: 5, Valid: true},
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
	})

	t.Run("topics page private title", func(t *testing.T) {
		origStore := core.Store
		origName := core.SessionName
		core.Store = sessions.NewCookieStore([]byte("test"))
		core.SessionName = "test-session"
		defer func() {
			core.Store = origStore
			core.SessionName = origName
		}()

		qs := testhelpers.NewQuerierStub()
		qs.GetAllForumCategoriesFn = func(ctx context.Context, arg db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
			return []*db.Forumcategory{}, nil
		}
		qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 1,
				Title:                        sql.NullString{String: "Private chat with Bob", Valid: true},
				Handler:                      "private",
				Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			}, nil
		}
		qs.AdminListPrivateTopicParticipantsByTopicIDFn = func(ctx context.Context, arg sql.NullInt32) ([]*db.AdminListPrivateTopicParticipantsByTopicIDRow, error) {
			return []*db.AdminListPrivateTopicParticipantsByTopicIDRow{
				{Idusers: 1, Username: sql.NullString{String: "Alice", Valid: true}},
				{Idusers: 2, Username: sql.NullString{String: "Bob", Valid: true}},
			}, nil
		}
		qs.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextFn = func(ctx context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
			return []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{}, nil
		}
		qs.ListContentPublicLabelsFn = func(arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{}, nil
		}

		cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())
		cd.UserID = 1 // Set viewer ID to 1 (Alice)

		req := httptest.NewRequest(http.MethodGet, "/forum/topic/1", nil)
		req = mux.SetURLVars(req, map[string]string{"topic": "1"})
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		w := httptest.NewRecorder()
		TopicsPage(w, req)

		body := w.Body.String()
		if strings.Contains(body, "Private chat with") {
			t.Fatalf("unexpected conversion message in output: %q", body)
		}
		if strings.Contains(body, "Category:") {
			t.Fatalf("unexpected category heading: %q", body)
		}
		// Expect Bob in the title (Alice is viewer, so excluded)
		if strings.Contains(body, "Alice") {
			t.Fatalf("unexpected participant name (Alice) in title, got %q", body)
		}
		if !strings.Contains(body, "Bob") {
			t.Fatalf("expected participant name (Bob), got %q", body)
		}
	})

	t.Run("thread page private sets title", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Comments:               sql.NullInt32{},
			Lastaddition:           sql.NullTime{},
			Locked:                 sql.NullBool{},
		}
		queries.GetForumTopicByIdForUserReturns = &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 1,
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{},
			Description:                  sql.NullString{},
			Threads:                      sql.NullInt32{},
			Comments:                     sql.NullInt32{},
			Lastaddition:                 sql.NullTime{},
			Handler:                      "private",
		}
		queries.AdminListPrivateTopicParticipantsByTopicIDReturns = []*db.AdminListPrivateTopicParticipantsByTopicIDRow{
			{Idusers: 2, Username: sql.NullString{String: "Bob", Valid: true}},
		}
		queries.GetCommentsByThreadIdForUserReturns = []*db.GetCommentsByThreadIdForUserRow{}

		origStore := core.Store
		origName := core.SessionName
		core.Store = sessions.NewCookieStore([]byte("test"))
		core.SessionName = "test-session"
		defer func() {
			core.Store = origStore
			core.SessionName = origName
		}()

		req := httptest.NewRequest("GET", "/private/topic/1/thread/1", nil)
		req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
		sess, _ := core.Store.New(req, core.SessionName)
		ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)

		cfg := config.NewRuntimeConfig()
		cd := common.NewCoreData(ctx, queries, cfg)
		cd.ShareSignKey = "secret"
		cd.SetCurrentSection("privateforum")
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		ThreadPageWithBasePath(rr, req, "/private")
		if cd.PageTitle == "" {
			t.Fatalf("page title not set")
		}
	})
}

func setupRequest(t *testing.T, queries db.Querier, path string, vars map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, vars)
	rr := httptest.NewRecorder()
	return req, rr
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

func (f *forumTopicPageQuerierFake) GetPermissionsByUserID(_ context.Context, idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
	f.record(fmt.Sprintf("GetPermissionsByUserID:%d", idusers))
	return nil, nil
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

func TestThreadPageTitle(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		// Setup
		queries := testhelpers.NewQuerierStub()
		queries.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Comments:               sql.NullInt32{},
			Lastaddition:           sql.NullTime{},
			Locked:                 sql.NullBool{},
		}
		queries.GetForumTopicByIdForUserReturns = &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 1,
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "My Topic", Valid: true},
			Description:                  sql.NullString{},
			Threads:                      sql.NullInt32{},
			Comments:                     sql.NullInt32{},
			Lastaddition:                 sql.NullTime{},
			Handler:                      "",
		}
		queries.GetForumCategoryByIdReturns = &db.Forumcategory{
			Idforumcategory: 1,
			Title:           sql.NullString{String: "My Category", Valid: true},
		}
		queries.GetCommentsByThreadIdForUserReturns = []*db.GetCommentsByThreadIdForUserRow{
			{
				Idcomments:    1,
				ForumthreadID: 1,
				Text:          sql.NullString{String: "This is the first post of the thread.", Valid: true},
			},
		}
		queries.ListContentPublicLabelsFn = func(arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{}, nil
		}
		queries.ListContentPrivateLabelsFn = func(arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
			return []*db.ListContentPrivateLabelsRow{}, nil
		}

		origStore := core.Store
		origName := core.SessionName
		core.Store = sessions.NewCookieStore([]byte("test"))
		core.SessionName = "test-session"
		defer func() {
			core.Store = origStore
			core.SessionName = origName
		}()

		req := httptest.NewRequest("GET", "/forum/topic/1/thread/1", nil)
		req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
		sess, _ := core.Store.New(req, core.SessionName)
		ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)

		cfg := config.NewRuntimeConfig()
		cd := common.NewCoreData(ctx, queries, cfg)
		cd.ShareSignKey = "secret"
		cd.SetCurrentSection("forum")
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		// Page execution
		rr := httptest.NewRecorder()
		ThreadPageWithBasePath(rr, req, "/forum")

		// Head data test
		expectedSnippet := "This is the first post..."
		expectedTitle := fmt.Sprintf("%s - %s - %s - Forum", expectedSnippet, "My Topic", "My Category")
		if cd.PageTitle != expectedTitle {
			t.Errorf("expected page title %q, got %q", expectedTitle, cd.PageTitle)
		}
	})
}

func TestTopicPageTitle(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		// Setup
		queries := testhelpers.NewQuerierStub()
		queries.GetForumTopicByIdForUserReturns = &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 1,
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "My Topic", Valid: true},
			Description:                  sql.NullString{},
			Threads:                      sql.NullInt32{},
			Comments:                     sql.NullInt32{},
			Lastaddition:                 sql.NullTime{},
			Handler:                      "",
		}
		queries.GetAllForumCategoriesReturns = []*db.Forumcategory{
			{
				Idforumcategory: 1,
				Title:           sql.NullString{String: "My Category", Valid: true},
			},
		}
		queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextFn = func(ctx context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
			return []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{}, nil
		}
		queries.ListContentPublicLabelsFn = func(arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{}, nil
		}

		origStore := core.Store
		origName := core.SessionName
		core.Store = sessions.NewCookieStore([]byte("test"))
		core.SessionName = "test-session"
		defer func() {
			core.Store = origStore
			core.SessionName = origName
		}()

		req := httptest.NewRequest("GET", "/forum/topic/1", nil)
		req = mux.SetURLVars(req, map[string]string{"topic": "1"})
		sess, _ := core.Store.New(req, core.SessionName)
		ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)

		cfg := config.NewRuntimeConfig()
		cd := common.NewCoreData(ctx, queries, cfg)
		cd.ShareSignKey = "secret"
		cd.SetCurrentSection("forum")
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		// Page execution
		rr := httptest.NewRecorder()
		TopicsPageWithBasePath(rr, req, "/forum")

		// Head data test
		expectedTitle := "My Topic - My Category - Forum"
		if cd.PageTitle != expectedTitle {
			t.Errorf("expected page title %q, got %q", expectedTitle, cd.PageTitle)
		}
	})
}
