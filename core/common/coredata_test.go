package common

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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
	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	cd := NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})
	ctx = context.WithValue(ctx, ContextValues("coreData"), cd)
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

	ctx := context.WithValue(context.Background(), ContextValues("queries"), queries)
	cd := NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})

	if _, err := cd.WritingCategories(); err != nil {
		t.Fatalf("WritingCategories: %v", err)
	}
	if _, err := cd.WritingCategories(); err != nil {
		t.Fatalf("WritingCategories second call: %v", err)
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

	ctx := context.WithValue(context.Background(), ContextValues("queries"), queries)
	cd := NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})

	req := httptest.NewRequest("GET", "/", nil).WithContext(context.WithValue(ctx, ContextValues("coreData"), cd))

	if _, err := cd.LatestWritings(req); err != nil {
		t.Fatalf("LatestWritings: %v", err)
	}
	if _, err := cd.LatestWritings(req); err != nil {
		t.Fatalf("LatestWritings second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestImageBoardsLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"idimageboard", "imageboard_idimageboard", "title", "description", "approval_required"}).
		AddRow(1, 0, sql.NullString{String: "t", Valid: true}, sql.NullString{String: "d", Valid: true}, true)

	mock.ExpectQuery("SELECT b.idimageboard").WithArgs(int32(1), int32(0), sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(rows)

	ctx := context.WithValue(context.Background(), ContextValues("queries"), queries)
	cd := NewCoreData(ctx, queries)
	cd.UserID = 1

	if _, err := cd.ImageBoards(0); err != nil {
		t.Fatalf("ImageBoards: %v", err)
	}
	if _, err := cd.ImageBoards(0); err != nil {
		t.Fatalf("ImageBoards second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestImageBoardPostsLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "username", "comments"}).
		AddRow(1, 0, 1, 2, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullString{}, 0, true, sql.NullTime{}, sql.NullString{}, sql.NullInt32{})

	mock.ExpectQuery("SELECT i.idimagepost").WithArgs(int32(1), int32(2), sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(rows)

	ctx := context.WithValue(context.Background(), ContextValues("queries"), queries)
	cd := NewCoreData(ctx, queries)
	cd.UserID = 1

	if _, err := cd.ImageBoardPosts(2); err != nil {
		t.Fatalf("ImageBoardPosts: %v", err)
	}
	if _, err := cd.ImageBoardPosts(2); err != nil {
		t.Fatalf("ImageBoardPosts second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
