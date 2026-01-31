package common

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestPrivateForumBreadcrumbUsesDisplayTitle(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	topicRows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "deleted_at", "LastPosterUsername"}).
		AddRow(1, 0, 0, nil, sql.NullString{String: "Private chat with Hidden", Valid: true}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, "private", sql.NullTime{}, sql.NullString{})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(topicRows)

	participantRows := sqlmock.NewRows([]string{"idusers", "username"}).
		AddRow(2, sql.NullString{String: "Alice", Valid: true}).
		AddRow(3, sql.NullString{String: "Bob", Valid: true})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, u.username")).
		WithArgs(sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(participantRows)

	queries := db.New(conn)
	cd := NewTestCoreData(t, queries)
	cd.SetCurrentSection("privateforum")
	cd.ForumBasePath = "/private"
	cd.UserID = 1
	cd.currentTopicID = 1

	crumbs, err := cd.forumBreadcrumbs()
	if err != nil {
		t.Fatalf("forumBreadcrumbs error: %v", err)
	}
	if len(crumbs) < 2 {
		t.Fatalf("expected >=2 crumbs, got %v", crumbs)
	}
	if crumbs[0].Title != "Private" {
		t.Fatalf("unexpected root crumb title: %v", crumbs[0].Title)
	}
	if crumbs[1].Title != "Alice, Bob" {
		t.Fatalf("unexpected crumb title: %v", crumbs[1].Title)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}
