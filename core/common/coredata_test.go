package common_test

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/consts"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCoreDataLatestNewsLazy(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id", "language_id",
		"users_idusers", "news", "occurred", "timezone", "comments",
	}).AddRow("w", 1, 1, 0, 1, 1, "a", now, time.Local.String(), 0)

	mock.ExpectQuery("SELECT u.username").WithArgs(int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	_ = req.WithContext(ctx)

	if _, err := cd.LatestNews(); err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if _, err := cd.LatestNews(); err != nil {
		t.Fatalf("LatestNews second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpdateFAQQuestion(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	cfg := config.NewRuntimeConfig()
	queries := db.New(conn)
	mock.ExpectExec("UPDATE faq").
		WithArgs(sql.NullString{String: "a", Valid: true}, sql.NullString{String: "q", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}, int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO faq_revisions").
		WithArgs(int32(1), int32(3), sql.NullString{String: "q", Valid: true}, sql.NullString{String: "a", Valid: true}, sql.NullString{String: cfg.Timezone, Valid: true}, sql.NullInt32{Int32: 3, Valid: true}, int32(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cd := common.NewTestCoreData(t, queries)
	common.WithConfig(cfg)(cd)
	if err := cd.UpdateFAQQuestion("q", "a", 2, 1, 3); err != nil {
		t.Fatalf("UpdateFAQQuestion: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritingCategoriesLazy(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "b")

	mock.ExpectQuery("SELECT wc.idwritingcategory").WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "category", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1

	if _, err := cd.VisibleWritingCategories(); err != nil {
		t.Fatalf("WritingCategories: %v", err)
	}
	if _, err := cd.VisibleWritingCategories(); err != nil {
		t.Fatalf("WritingCategories second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNewsAnnouncementCaching(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	now := time.Now()
	annRows := sqlmock.NewRows([]string{"id", "site_news_id", "active", "created_at"}).
		AddRow(1, 1, true, now)

	mock.ExpectQuery("SELECT id, site_news_id, active, created_at").WithArgs(int32(1)).WillReturnRows(annRows)

	cd := common.NewTestCoreData(t, queries)

	if cd.NewsAnnouncement(1) == nil {
		t.Fatalf("NewsAnnouncement returned nil")
	}
	if cd.NewsAnnouncement(1) == nil {
		t.Fatalf("NewsAnnouncement second returned nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNewsAnnouncementError(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	mock.ExpectQuery("SELECT id, site_news_id, active, created_at").WithArgs(int32(1)).WillReturnError(sql.ErrConnDone)

	cd := common.NewTestCoreData(t, queries)

	if cd.NewsAnnouncement(1) != nil {
		t.Fatalf("NewsAnnouncement expected nil on error")
	}
	if cd.NewsAnnouncement(1) != nil {
		t.Fatalf("NewsAnnouncement second expected nil on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPublicWritingsLazy(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"idwriting", "users_idusers", "forumthread_id", "language_id", "writing_category_id", "title", "published", "timezone", "writing", "abstract", "private", "deleted_at", "last_index", "Username", "Comments"}).
		AddRow(1, 1, 0, 1, 0, "t", now, time.Local.String(), "w", "a", false, now, now, "u", 0)

	mock.ExpectQuery("SELECT w.idwriting").WithArgs(int32(1), int32(0), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	rows2 := sqlmock.NewRows([]string{"idwriting", "users_idusers", "forumthread_id", "language_id", "writing_category_id", "title", "published", "timezone", "writing", "abstract", "private", "deleted_at", "last_index", "Username", "Comments"}).
		AddRow(2, 1, 0, 1, 1, "t2", now, time.Local.String(), "w2", "a2", false, now, now, "u", 0)

	mock.ExpectQuery("SELECT w.idwriting").WithArgs(int32(1), int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows2)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "see", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCoreDataLatestWritingsLazy(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"idwriting", "users_idusers", "forumthread_id", "language_id",
		"writing_category_id", "title", "published", "timezone", "writing", "abstract",
		"private", "deleted_at", "last_index",
	}).AddRow(1, 1, 0, 1, 1, "t", now, time.Local.String(), "w", "a", nil, nil, now)

	mock.ExpectQuery("SELECT w.idwriting").WithArgs(int32(15), int32(0)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	offset, _ := strconv.Atoi(req.URL.Query().Get("offset"))
	if _, err := cd.LatestWritings(common.WithWritingsOffset(int32(offset))); err != nil {
		t.Fatalf("LatestWritings: %v", err)
	}
	if _, err := cd.LatestWritings(common.WithWritingsOffset(int32(offset))); err != nil {
		t.Fatalf("LatestWritings second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBloggersLazy(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery("SELECT u.username").
		WithArgs(int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(16), int32(0)).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewTestCoreData(t, queries)
	common.WithConfig(cfg)(cd)
	cd.UserID = 1
	req = req.WithContext(ctx)

	if _, err := cd.Bloggers(req); err != nil {
		t.Fatalf("Bloggers: %v", err)
	}
	if _, err := cd.Bloggers(req); err != nil {
		t.Fatalf("Bloggers second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritersLazy(t *testing.T) {

	cfg := config.NewRuntimeConfig()

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery("SELECT u.username").
		WithArgs(int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(16), int32(0)).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewTestCoreData(t, queries)
	common.WithConfig(cfg)(cd)
	cd.UserID = 1
	req = req.WithContext(ctx)

	if _, err := cd.Writers(req); err != nil {
		t.Fatalf("Writers: %v", err)
	}
	if _, err := cd.Writers(req); err != nil {
		t.Fatalf("Writers second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBlogListLazy(t *testing.T) {

	cfg := config.NewRuntimeConfig()

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_id", "blog", "written", "timezone", "username", "comments", "is_owner"}).
		AddRow(1, nil, 1, 0, "b", now, time.Local.String(), "bob", 0, true)
	mock.ExpectQuery("SELECT b.idblogs").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	cd := common.NewTestCoreData(t, queries)
	common.WithConfig(cfg)(cd)
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.UserID = 1

	if _, err := cd.BlogList(); err != nil {
		t.Fatalf("BlogList: %v", err)
	}
	if _, err := cd.BlogList(); err != nil {
		t.Fatalf("BlogList second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBlogListForSelectedAuthorLazy(t *testing.T) {

	cfg := config.NewRuntimeConfig()

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_id", "blog", "written", "timezone", "username", "comments", "is_owner"}).
		AddRow(1, nil, 1, 0, "b", now, time.Local.String(), "bob", 0, true)
	mock.ExpectQuery("SELECT b.idblogs").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	cd := common.NewTestCoreData(t, queries)
	common.WithConfig(cfg)(cd)
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.UserID = 1
	cd.SetCurrentProfileUserID(1)

	if _, err := cd.BlogListForSelectedAuthor(); err != nil {
		t.Fatalf("BlogListForSelectedAuthor: %v", err)
	}
	if _, err := cd.BlogListForSelectedAuthor(); err != nil {
		t.Fatalf("BlogListForSelectedAuthor second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSelectedQuestionFromCategory(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)

	row := sqlmock.NewRows([]string{"id", "category_id", "language_id", "author_id", "answer", "question"}).
		AddRow(1, 2, 0, 0, sql.NullString{}, sql.NullString{})
	mock.ExpectQuery("SELECT id, category_id").WithArgs(int32(1)).WillReturnRows(row)
	mock.ExpectExec("UPDATE faq SET deleted_at").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SelectedQuestionFromCategory(1, 2); err != nil {
		t.Fatalf("SelectedQuestionFromCategory: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSelectedQuestionFromCategoryWrongCategory(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)

	row := sqlmock.NewRows([]string{"id", "category_id", "language_id", "author_id", "answer", "question"}).
		AddRow(1, 3, 0, 0, sql.NullString{}, sql.NullString{})
	mock.ExpectQuery("SELECT id, category_id").WithArgs(int32(1)).WillReturnRows(row)

	if err := cd.SelectedQuestionFromCategory(1, 2); err == nil {
		t.Fatalf("expected error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSelectedThreadCanReply(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	cd.SetCurrentSection("forum")
	threadID, topicID := int32(3), int32(2)
	cd.SetCurrentThreadAndTopic(threadID, topicID)

	rows := sqlmock.NewRows([]string{
		"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked",
	}).AddRow(threadID, 0, 0, topicID, nil, time.Now(), nil)
	mock.ExpectQuery("SELECT th.idforumthread").WithArgs(
		int32(1),
		threadID,
		int32(1),
		int32(1),
		"forum",
		sql.NullString{String: "topic", Valid: true},
		sql.NullInt32{Int32: topicID, Valid: true},
		sql.NullInt32{Int32: 1, Valid: true},
	).WillReturnRows(rows)

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSelectedThreadCanReplyPrivateForum(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	cd.SetCurrentSection("privateforum")
	threadID, topicID := int32(3), int32(2)
	cd.SetCurrentThreadAndTopic(threadID, topicID)

	rows := sqlmock.NewRows([]string{
		"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked",
	}).AddRow(threadID, 0, 0, topicID, nil, time.Now(), nil)
	mock.ExpectQuery("SELECT th.idforumthread").WithArgs(
		int32(1),
		threadID,
		int32(1),
		int32(1),
		"privateforum",
		sql.NullString{String: "topic", Valid: true},
		sql.NullInt32{Int32: topicID, Valid: true},
		sql.NullInt32{Int32: 1, Valid: true},
	).WillReturnRows(rows)

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSelectedThreadCanReplyGrantFallback(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	cd.SetCurrentSection("blogs")
	threadID, blogID := int32(5), int32(7)
	cd.SetCurrentThreadAndTopic(threadID, 0)
	cd.SetCurrentBlog(blogID)

	mock.ExpectQuery("SELECT th.idforumthread").WithArgs(
		int32(1),
		threadID,
		int32(1),
		int32(1),
		"blogs",
		sql.NullString{String: "entry", Valid: true},
		sql.NullInt32{Int32: blogID, Valid: true},
		sql.NullInt32{Int32: 1, Valid: true},
	).WillReturnError(sql.ErrNoRows)

	grantRows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery("SELECT 1 FROM grants g").WithArgs(
		int32(1),
		"blogs",
		sql.NullString{String: "entry", Valid: true},
		"reply",
		sql.NullInt32{Int32: blogID, Valid: true},
		sql.NullInt32{Int32: 1, Valid: true},
	).WillReturnRows(grantRows)

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSelectedThreadCanReplyGrantFallbackNoThread(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	cd.SetCurrentSection("blogs")
	blogID := int32(7)
	cd.SetCurrentBlog(blogID)

	grantRows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery("SELECT 1 FROM grants g").WithArgs(
		int32(1),
		"blogs",
		sql.NullString{String: "entry", Valid: true},
		"reply",
		sql.NullInt32{Int32: blogID, Valid: true},
		sql.NullInt32{Int32: 1, Valid: true},
	).WillReturnRows(grantRows)

	if !cd.SelectedThreadCanReply() {
		t.Fatalf("SelectedThreadCanReply() = false; want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
