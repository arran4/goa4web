package db

import (
	"context"
	"testing"

	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
)

func TestQueries_AdminListUsersFiltered(t *testing.T) {
	q := newAdminUsersQuerier([]string{"idusers", "email", "username"}, []driver.Value{int32(1), "bob@example.com", "bob"})

	res, err := q.AdminListUsersFiltered(context.Background(), AdminListUsersFilteredParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListUsersFiltered: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Email.String != "bob@example.com" || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestQueries_AdminSearchUsersFiltered(t *testing.T) {
	q := newAdminUsersQuerier([]string{"idusers", "email", "username"}, []driver.Value{int32(1), "bob@example.com", "bob"})

	res, err := q.AdminSearchUsersFiltered(context.Background(), AdminSearchUsersFilteredParams{Query: "bob", Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminSearchUsersFiltered: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Email.String != "bob@example.com" || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}
}

func newAdminUsersQuerier(columns []string, rows ...[]driver.Value) *Queries {
	conn := &adminUsersConnector{columns: columns, rows: rows}
	db := sql.OpenDB(conn)

	return New(db)
}

type adminUsersConnector struct {
	columns []string
	rows    [][]driver.Value
}

func (c *adminUsersConnector) Connect(context.Context) (driver.Conn, error) {
	return &adminUsersConn{columns: c.columns, rows: c.rows}, nil
}

func (c *adminUsersConnector) Driver() driver.Driver { return adminUsersDriver{} }

type adminUsersDriver struct{}

func (adminUsersDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use Connector") }

type adminUsersConn struct {
	columns []string
	rows    [][]driver.Value
}

func (c *adminUsersConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *adminUsersConn) Close() error { return nil }

func (c *adminUsersConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

func (c *adminUsersConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return &staticRows{columns: c.columns, rows: append([][]driver.Value(nil), c.rows...)}, nil
}

func (c *adminUsersConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("not implemented")
}

type staticRows struct {
	columns []string
	rows    [][]driver.Value
	pos     int
}

func (r *staticRows) Columns() []string { return r.columns }

func (r *staticRows) Close() error { return nil }

func (r *staticRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}
