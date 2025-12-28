package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_AdminListUsersFiltered(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	query := "SELECT u.idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, u.username FROM users u ORDER BY u.idusers LIMIT ? OFFSET ?"
	rows := sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "bob@example.com", "bob")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminListUsersFiltered(context.Background(), AdminListUsersFilteredParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListUsersFiltered: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Email.String != "bob@example.com" || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminSearchUsersFiltered(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	query := "SELECT u.idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, u.username FROM users u WHERE (LOWER(u.username) LIKE LOWER(?) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(?)) ORDER BY u.idusers LIMIT ? OFFSET ?"
	rows := sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "bob@example.com", "bob")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("%bob%", "%bob%", int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.AdminSearchUsersFiltered(context.Background(), AdminSearchUsersFilteredParams{Query: "bob", Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminSearchUsersFiltered: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Email.String != "bob@example.com" || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
