package common_test

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestCoreDataLatestNewsLazy(t *testing.T) {
	queries := testutil.NewNewsQuerier(t)
	queries.AllowGrants()
	now := time.Now()
	queries.Posts = []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow{{
		Writername: sql.NullString{String: "w", Valid: true},
		Writerid:   sql.NullInt32{Int32: 1, Valid: true},
		Idsitenews: 1,
		Occurred:   sql.NullTime{Time: now, Valid: true},
		Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
	}}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	_ = req.WithContext(ctx)

	if _, err := cd.LatestNews(); err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if _, err := cd.LatestNews(); err != nil {
		t.Fatalf("LatestNews second call: %v", err)
	}

}

func TestUpdateFAQQuestion(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	queries := testutil.NewFAQQuerier(t)

	cd := common.NewCoreData(context.Background(), queries, cfg, common.WithPreference(&db.Preference{PageSize: 15}))
	if err := cd.UpdateFAQQuestion("q", "a", 2, 1, 3); err != nil {
		t.Fatalf("UpdateFAQQuestion: %v", err)
	}
	if len(queries.Updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(queries.Updated))
	}
	if queries.Updated[0].ID != 1 {
		t.Fatalf("AdminUpdateFAQQuestionAnswer ID %d, want 1", queries.Updated[0].ID)
	}
	if len(queries.Revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(queries.Revisions))
	}
	rev := queries.Revisions[0]
	if rev.FaqID != 1 || rev.UsersIdusers != 3 {
		t.Fatalf("InsertFAQRevisionForUser IDs %+v", rev)
	}
	if rev.Timezone.String != cfg.Timezone {
		t.Fatalf("InsertFAQRevisionForUser timezone %q, want %q", rev.Timezone.String, cfg.Timezone)
	}

}

func TestWritingCategoriesLazy(t *testing.T) {
	queries := testutil.NewWritingCategoriesQuerier(t)
	queries.AllowGrants()
	queries.Categories = []*db.WritingCategory{{
		Idwritingcategory: 1,
		WritingCategoryID: sql.NullInt32{Int32: 0, Valid: false},
		Title:             sql.NullString{String: "a", Valid: true},
		Description:       sql.NullString{String: "b", Valid: true},
	}}

	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1

	if _, err := cd.VisibleWritingCategories(); err != nil {
		t.Fatalf("WritingCategories: %v", err)
	}
	if _, err := cd.VisibleWritingCategories(); err != nil {
		t.Fatalf("WritingCategories second call: %v", err)
	}

}

func TestNewsAnnouncementCaching(t *testing.T) {
	queries := testutil.NewAnnouncementQuerier(t)
	now := time.Now()
	queries.Announcement = &db.SiteAnnouncement{ID: 1, SiteNewsID: 1, Active: true, CreatedAt: now}

	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithPreference(&db.Preference{PageSize: 15}))

	if cd.NewsAnnouncement(1) == nil {
		t.Fatalf("NewsAnnouncement returned nil")
	}
	if cd.NewsAnnouncement(1) == nil {
		t.Fatalf("NewsAnnouncement second returned nil")
	}

}

func TestNewsAnnouncementError(t *testing.T) {
	queries := testutil.NewAnnouncementQuerier(t)
	queries.Err = sql.ErrConnDone

	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithPreference(&db.Preference{PageSize: 15}))

	if cd.NewsAnnouncement(1) != nil {
		t.Fatalf("NewsAnnouncement expected nil on error")
	}
	if cd.NewsAnnouncement(1) != nil {
		t.Fatalf("NewsAnnouncement second expected nil on error")
	}

}

func TestPublicWritingsLazy(t *testing.T) {
	queries := testutil.NewWritingsQuerier(t)
	queries.AllowGrants()
	now := time.Now()
	queries.PublicRowsByCategory = map[int32][]*db.ListPublicWritingsInCategoryForListerRow{
		0: {{
			Idwriting:         1,
			WritingCategoryID: 0,
			Title:             sql.NullString{String: "t", Valid: true},
			Published:         sql.NullTime{Time: now, Valid: true},
			Timezone:          sql.NullString{String: time.Local.String(), Valid: true},
		}},
		1: {{
			Idwriting:         2,
			WritingCategoryID: 1,
			Title:             sql.NullString{String: "t2", Valid: true},
			Published:         sql.NullTime{Time: now, Valid: true},
			Timezone:          sql.NullString{String: time.Local.String(), Valid: true},
		}},
	}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if _, err := cd.PublicWritings(0, req); err != nil {
		t.Fatalf("PublicWritings: %v", err)
	}
	if _, err := cd.PublicWritings(0, req); err != nil {
		t.Fatalf("PublicWritings second call: %v", err)
	}
	if _, err := cd.PublicWritings(1, req); err != nil {
		t.Fatalf("PublicWritings other category: %v", err)
	}
	if _, err := cd.PublicWritings(1, req); err != nil {
		t.Fatalf("PublicWritings other category second call: %v", err)
	}

}

func TestCoreDataLatestWritingsLazy(t *testing.T) {
	queries := testutil.NewWritingsQuerier(t)
	queries.AllowGrants()
	now := time.Now()
	queries.Writings = []*db.Writing{{
		Idwriting:         1,
		WritingCategoryID: 1,
		Title:             sql.NullString{String: "t", Valid: true},
		Published:         sql.NullTime{Time: now, Valid: true},
		Timezone:          sql.NullString{String: time.Local.String(), Valid: true},
	}}

	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/", nil).WithContext(context.WithValue(ctx, consts.KeyCoreData, cd))
	offset, _ := strconv.Atoi(req.URL.Query().Get("offset"))
	if _, err := cd.LatestWritings(common.WithWritingsOffset(int32(offset))); err != nil {
		t.Fatalf("LatestWritings: %v", err)
	}
	if _, err := cd.LatestWritings(common.WithWritingsOffset(int32(offset))); err != nil {
		t.Fatalf("LatestWritings second call: %v", err)
	}

}

func TestBloggersLazy(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	queries := testutil.NewBlogQuerier(t)
	queries.Bloggers = []*db.ListBloggersForListerRow{{Username: sql.NullString{String: "bob", Valid: true}, Count: 2}}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, cfg, common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1
	req = req.WithContext(ctx)

	if _, err := cd.Bloggers(req); err != nil {
		t.Fatalf("Bloggers: %v", err)
	}
	if _, err := cd.Bloggers(req); err != nil {
		t.Fatalf("Bloggers second call: %v", err)
	}

}

func TestWritersLazy(t *testing.T) {

	cfg := config.NewRuntimeConfig()

	queries := testutil.NewBlogQuerier(t)
	queries.Writers = []*db.ListWritersForListerRow{{Username: sql.NullString{String: "bob", Valid: true}, Count: 2}}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, cfg, common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1
	req = req.WithContext(ctx)

	if _, err := cd.Writers(req); err != nil {
		t.Fatalf("Writers: %v", err)
	}
	if _, err := cd.Writers(req); err != nil {
		t.Fatalf("Writers second call: %v", err)
	}

}

func TestBlogListLazy(t *testing.T) {

	cfg := config.NewRuntimeConfig()

	queries := testutil.NewBlogQuerier(t)
	queries.AllowGrants()
	now := time.Now()
	queries.BlogEntries = []*db.ListBlogEntriesForListerRow{{
		Idblogs:  1,
		Blog:     sql.NullString{String: "b", Valid: true},
		Written:  now,
		Timezone: sql.NullString{String: time.Local.String(), Valid: true},
		Username: sql.NullString{String: "bob", Valid: true},
		Comments: 0,
		IsOwner:  true,
	}}

	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, cfg, common.WithUserRoles([]string{"administrator"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1

	if _, err := cd.BlogList(); err != nil {
		t.Fatalf("BlogList: %v", err)
	}
	if _, err := cd.BlogList(); err != nil {
		t.Fatalf("BlogList second call: %v", err)
	}

}

func TestBlogListForSelectedAuthorLazy(t *testing.T) {

	cfg := config.NewRuntimeConfig()

	queries := testutil.NewBlogQuerier(t)
	queries.AllowGrants()
	now := time.Now()
	queries.BlogEntriesByAuthor = []*db.ListBlogEntriesByAuthorForListerRow{{
		Idblogs:  1,
		Blog:     sql.NullString{String: "b", Valid: true},
		Written:  now,
		Timezone: sql.NullString{String: time.Local.String(), Valid: true},
		Username: sql.NullString{String: "bob", Valid: true},
		Comments: 0,
		IsOwner:  true,
	}}

	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, cfg, common.WithUserRoles([]string{"administrator"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1
	cd.SetCurrentProfileUserID(1)

	if _, err := cd.BlogListForSelectedAuthor(); err != nil {
		t.Fatalf("BlogListForSelectedAuthor: %v", err)
	}
	if _, err := cd.BlogListForSelectedAuthor(); err != nil {
		t.Fatalf("BlogListForSelectedAuthor second call: %v", err)
	}

}

func TestSelectedQuestionFromCategory(t *testing.T) {
	queries := testutil.NewFAQQuerier(t)
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())

	queries.FAQ = &db.Faq{ID: 1, CategoryID: sql.NullInt32{Int32: 2, Valid: true}}

	if err := cd.SelectedQuestionFromCategory(1, 2); err != nil {
		t.Fatalf("SelectedQuestionFromCategory: %v", err)
	}
	if len(queries.DeletedIDs) != 1 || queries.DeletedIDs[0] != 1 {
		t.Fatalf("AdminDeleteFAQ ids %+v, want [1]", queries.DeletedIDs)
	}

}

func TestSelectedQuestionFromCategoryWrongCategory(t *testing.T) {
	queries := testutil.NewFAQQuerier(t)
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())

	queries.FAQ = &db.Faq{ID: 1, CategoryID: sql.NullInt32{Int32: 3, Valid: true}}

	if err := cd.SelectedQuestionFromCategory(1, 2); err == nil {
		t.Fatalf("expected error")
	}

}

func TestSelectedThreadCanReply(t *testing.T) {
	queries := testutil.NewThreadReplyQuerier(t)
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1
	cd.SetCurrentSection("forum")
	threadID, topicID := int32(3), int32(2)
	cd.SetCurrentThreadAndTopic(threadID, topicID)

	queries.Thread = &db.Forumthread{Idforumthread: threadID, ForumtopicIdforumtopic: topicID}

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

}

func TestSelectedThreadCanReplyPrivateForum(t *testing.T) {
	queries := testutil.NewThreadReplyQuerier(t)
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1
	cd.SetCurrentSection("privateforum")
	threadID, topicID := int32(3), int32(2)
	cd.SetCurrentThreadAndTopic(threadID, topicID)

	queries.Thread = &db.Forumthread{Idforumthread: threadID, ForumtopicIdforumtopic: topicID}

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

}

func TestSelectedThreadCanReplyGrantFallback(t *testing.T) {
	queries := testutil.NewThreadReplyQuerier(t)
	queries.AllowGrants()
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1
	cd.SetCurrentSection("blogs")
	threadID, blogID := int32(5), int32(7)
	cd.SetCurrentThreadAndTopic(threadID, 0)
	cd.SetCurrentBlog(blogID)

	queries.Err = sql.ErrNoRows

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

}

func TestSelectedThreadCanReplyGrantFallbackNoThread(t *testing.T) {
	queries := testutil.NewBaseQuerier(t)
	queries.AllowGrants()
	ctx := context.Background()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1
	cd.SetCurrentSection("blogs")
	blogID := int32(7)
	cd.SetCurrentBlog(blogID)

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

}
