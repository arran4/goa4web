package common_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	common "github.com/arran4/goa4web/core/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestCoreDataLatestNewsLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id", "language_idlanguage",
		"users_idusers", "news", "occurred", "comments",
	}).AddRow("w", 1, 1, 0, 1, 1, "a", now, 0)

	mock.ExpectQuery("SELECT u.username").WithArgs(int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT id, site_news_id, active, created_at").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})
	ctx = context.WithValue(ctx, common.ContextValues("coreData"), cd)
	req = req.WithContext(ctx)

	if _, err := cd.LatestNews(req); err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if _, err := cd.LatestNews(req); err != nil {
		t.Fatalf("LatestNews second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritingCategoriesLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "b")

	mock.ExpectQuery("SELECT wc.idwritingcategory").WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "category", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	ctx := context.WithValue(context.Background(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})

	if _, err := cd.VisibleWritingCategories(); err != nil {
		t.Fatalf("VisibleWritingCategories: %v", err)
	}
	if _, err := cd.VisibleWritingCategories(); err != nil {
		t.Fatalf("VisibleWritingCategories second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAnnouncementForNewsCaching(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	now := time.Now()
	annRows := sqlmock.NewRows([]string{"id", "site_news_id", "active", "created_at"}).
		AddRow(1, 1, true, now)

	mock.ExpectQuery("SELECT id, site_news_id, active, created_at").WithArgs(int32(1)).WillReturnRows(annRows)

	ctx := context.WithValue(context.Background(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)

	if _, err := cd.AnnouncementForNews(1); err != nil {
		t.Fatalf("AnnouncementForNews: %v", err)
	}
	if _, err := cd.AnnouncementForNews(1); err != nil {
		t.Fatalf("AnnouncementForNews second: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAnnouncementForNewsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	mock.ExpectQuery("SELECT id, site_news_id, active, created_at").WithArgs(int32(1)).WillReturnError(sql.ErrConnDone)

	ctx := context.WithValue(context.Background(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)

	if _, err := cd.AnnouncementForNews(1); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("AnnouncementForNews error=%v", err)
	}
	if _, err := cd.AnnouncementForNews(1); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("AnnouncementForNews second=%v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPublicWritingsLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"idwriting", "users_idusers", "forumthread_id", "language_idlanguage", "writing_category_id", "title", "published", "writing", "abstract", "private", "deleted_at", "Username", "Comments"}).
		AddRow(1, 1, 0, 1, 0, "t", now, "w", "a", false, now, "u", 0)

	mock.ExpectQuery("SELECT w.idwriting").WithArgs(int32(1), int32(0), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	rows2 := sqlmock.NewRows([]string{"idwriting", "users_idusers", "forumthread_id", "language_idlanguage", "writing_category_id", "title", "published", "writing", "abstract", "private", "deleted_at", "Username", "Comments"}).
		AddRow(2, 1, 0, 1, 1, "t2", now, "w2", "a2", false, now, "u", 0)

	mock.ExpectQuery("SELECT w.idwriting").WithArgs(int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows2)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "see", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})
	ctx = context.WithValue(ctx, common.ContextValues("coreData"), cd)
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"idwriting", "users_idusers", "forumthread_id", "language_idlanguage",
		"writing_category_id", "title", "published", "writing", "abstract",
		"private", "deleted_at",
	}).AddRow(1, 1, 0, 1, 1, "t", now, "w", "a", nil, nil)

	mock.ExpectQuery("SELECT w.idwriting").WithArgs(int32(15), int32(0)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	ctx := context.WithValue(context.Background(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})

	if _, err := cd.LatestWritings(); err != nil {
		t.Fatalf("LatestWritings: %v", err)
	}
	if _, err := cd.LatestWritings(); err != nil {
		t.Fatalf("LatestWritings second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBloggersLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery("SELECT u.username").
		WithArgs(int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(16), int32(0)).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery("SELECT u.username").
		WithArgs(int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(16), int32(0)).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
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
