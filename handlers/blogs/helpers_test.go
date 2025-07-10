package blogs

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func TestCurrentUserMayReply(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), hcommon.KeyQueries, q)
	ctx = context.WithValue(ctx, hcommon.KeyCoreData, &hcommon.CoreData{UserID: 1})
	req = req.WithContext(ctx)

	blog := &db.GetBlogEntryForUserByIdRow{ForumthreadID: 2}
	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(1), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic",
			"comments", "lastaddition", "locked", "LastPosterUsername", "seelevel", "level",
		}).AddRow(2, 0, 0, 1, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{Bool: false, Valid: true}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}))

	if !currentUserMayReply(req, blog) {
		t.Errorf("expected true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCurrentUserMayReplyLocked(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), hcommon.KeyQueries, q)
	ctx = context.WithValue(ctx, hcommon.KeyCoreData, &hcommon.CoreData{UserID: 1})
	req = req.WithContext(ctx)

	blog := &db.GetBlogEntryForUserByIdRow{ForumthreadID: 2}
	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(1), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic",
			"comments", "lastaddition", "locked", "LastPosterUsername", "seelevel", "level",
		}).AddRow(2, 0, 0, 1, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{Bool: true, Valid: true}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}))

	if currentUserMayReply(req, blog) {
		t.Errorf("expected false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCurrentUserMayReplyNoUser(t *testing.T) {
	sqldb, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), hcommon.KeyQueries, q)
	// no CoreData user
	req = req.WithContext(ctx)

	blog := &db.GetBlogEntryForUserByIdRow{ForumthreadID: 0}
	if currentUserMayReply(req, blog) {
		t.Errorf("expected false")
	}
}
