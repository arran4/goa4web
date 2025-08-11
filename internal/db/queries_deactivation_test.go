package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_AdminIsUserDeactivated(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"is_deactivated"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(adminIsUserDeactivated)).
		WithArgs(int32(1)).
		WillReturnRows(rows)

	res, err := q.AdminIsUserDeactivated(context.Background(), 1)
	if err != nil {
		t.Fatalf("AdminIsUserDeactivated: %v", err)
	}
	if !res {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListDeactivatedUsers(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@example.com", "alice")
	mock.ExpectQuery(regexp.QuoteMeta(adminListDeactivatedUsers)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListDeactivatedUsers(context.Background(), AdminListDeactivatedUsersParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminIsBlogDeactivated(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"is_deactivated"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(adminIsBlogDeactivated)).
		WithArgs(int32(2)).
		WillReturnRows(rows)

	res, err := q.AdminIsBlogDeactivated(context.Background(), 2)
	if err != nil {
		t.Fatalf("AdminIsBlogDeactivated: %v", err)
	}
	if !res {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListDeactivatedBlogs(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idblogs", "blog"}).AddRow(1, "entry")
	mock.ExpectQuery(regexp.QuoteMeta(adminListDeactivatedBlogs)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListDeactivatedBlogs(context.Background(), AdminListDeactivatedBlogsParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedBlogs: %v", err)
	}
	if len(res) != 1 || res[0].Idblogs != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminIsCommentDeactivated(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"is_deactivated"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(adminIsCommentDeactivated)).
		WithArgs(int32(3)).
		WillReturnRows(rows)

	res, err := q.AdminIsCommentDeactivated(context.Background(), 3)
	if err != nil {
		t.Fatalf("AdminIsCommentDeactivated: %v", err)
	}
	if !res {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListDeactivatedComments(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idcomments", "text"}).AddRow(1, "t")
	mock.ExpectQuery(regexp.QuoteMeta(adminListDeactivatedComments)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListDeactivatedComments(context.Background(), AdminListDeactivatedCommentsParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedComments: %v", err)
	}
	if len(res) != 1 || res[0].Idcomments != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminIsWritingDeactivated(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"is_deactivated"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(adminIsWritingDeactivated)).
		WithArgs(int32(4)).
		WillReturnRows(rows)

	res, err := q.AdminIsWritingDeactivated(context.Background(), 4)
	if err != nil {
		t.Fatalf("AdminIsWritingDeactivated: %v", err)
	}
	if !res {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListDeactivatedWritings(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idwriting", "title", "writing", "abstract", "private"}).
		AddRow(1, "t", "w", "a", true)
	mock.ExpectQuery(regexp.QuoteMeta(adminListDeactivatedWritings)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListDeactivatedWritings(context.Background(), AdminListDeactivatedWritingsParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedWritings: %v", err)
	}
	if len(res) != 1 || res[0].Idwriting != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminIsImagepostDeactivated(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"is_deactivated"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(adminIsImagepostDeactivated)).
		WithArgs(int32(5)).
		WillReturnRows(rows)

	res, err := q.AdminIsImagepostDeactivated(context.Background(), 5)
	if err != nil {
		t.Fatalf("AdminIsImagepostDeactivated: %v", err)
	}
	if !res {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListDeactivatedImageposts(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idimagepost", "description", "thumbnail", "fullimage"}).
		AddRow(1, "d", "t", "f")
	mock.ExpectQuery(regexp.QuoteMeta(adminListDeactivatedImageposts)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListDeactivatedImageposts(context.Background(), AdminListDeactivatedImagepostsParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedImageposts: %v", err)
	}
	if len(res) != 1 || res[0].Idimagepost != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminIsLinkDeactivated(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"is_deactivated"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(adminIsLinkDeactivated)).
		WithArgs(int32(6)).
		WillReturnRows(rows)

	res, err := q.AdminIsLinkDeactivated(context.Background(), 6)
	if err != nil {
		t.Fatalf("AdminIsLinkDeactivated: %v", err)
	}
	if !res {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListDeactivatedLinks(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idlinker", "title", "url", "description"}).
		AddRow(1, "t", "u", "d")
	mock.ExpectQuery(regexp.QuoteMeta(adminListDeactivatedLinks)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListDeactivatedLinks(context.Background(), AdminListDeactivatedLinksParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedLinks: %v", err)
	}
	if len(res) != 1 || res[0].ID != 1 {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
