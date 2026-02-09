package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
)

// mockConn implements driver.Conn and driver.ExecerContext
type mockConn struct {
	lastQuery string
	lastArgs  []driver.Value
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *mockConn) Close() error { return nil }

func (c *mockConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

func (c *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.lastQuery = query
	c.lastArgs = make([]driver.Value, len(args))
	for i, arg := range args {
		c.lastArgs[i] = arg.Value
	}
	return mockResult{}, nil
}

type mockResult struct{}

func (r mockResult) LastInsertId() (int64, error) { return 0, nil }
func (r mockResult) RowsAffected() (int64, error) { return 1, nil }

// mockConnector implements driver.Connector
type mockConnector struct {
	conn *mockConn
}

func (c *mockConnector) Connect(context.Context) (driver.Conn, error) {
	return c.conn, nil
}

func (c *mockConnector) Driver() driver.Driver { return nil }

func newMockQueries() (*Queries, *mockConn) {
	conn := &mockConn{}
	db := sql.OpenDB(&mockConnector{conn: conn})
	return New(db), conn
}

func TestBatchInsertSubscriptions(t *testing.T) {
	q, conn := newMockQueries()
	ctx := context.Background()

	params := []InsertSubscriptionParams{
		{UsersIdusers: 1, Pattern: "p1", Method: "m1"},
		{UsersIdusers: 1, Pattern: "p2", Method: "m2"},
	}

	err := q.BatchInsertSubscriptions(ctx, params)
	if err != nil {
		t.Fatalf("BatchInsertSubscriptions failed: %v", err)
	}

	expectedQueryStart := "INSERT INTO subscriptions (users_idusers, pattern, method) VALUES "
	if !strings.HasPrefix(conn.lastQuery, expectedQueryStart) {
		t.Errorf("unexpected query start: got %s", conn.lastQuery)
	}

	expectedPlaceholders := "(?, ?, ?),(?, ?, ?)"
	if !strings.Contains(conn.lastQuery, expectedPlaceholders) {
		t.Errorf("unexpected placeholders: got %s", conn.lastQuery)
	}

	if len(conn.lastArgs) != 6 {
		t.Errorf("unexpected args count: got %d, want 6", len(conn.lastArgs))
	}

	// Verify args - driver converts integers to int64
	expectedArgs := []interface{}{int64(1), "p1", "m1", int64(1), "p2", "m2"}
	for i, arg := range conn.lastArgs {
		if arg != expectedArgs[i] {
			t.Errorf("arg %d mismatch: got %v, want %v", i, arg, expectedArgs[i])
		}
	}
}

func TestBatchDeleteSubscriptions(t *testing.T) {
	q, conn := newMockQueries()
	ctx := context.Background()

	params := []DeleteSubscriptionForSubscriberParams{
		{SubscriberID: 1, Pattern: "p1", Method: "m1"},
		{SubscriberID: 1, Pattern: "p2", Method: "m2"},
	}

	err := q.BatchDeleteSubscriptions(ctx, params)
	if err != nil {
		t.Fatalf("BatchDeleteSubscriptions failed: %v", err)
	}

	expectedQueryStart := "DELETE FROM subscriptions WHERE (users_idusers, pattern, method) IN ("
	if !strings.HasPrefix(conn.lastQuery, expectedQueryStart) {
		t.Errorf("unexpected query start: got %s", conn.lastQuery)
	}

	expectedPlaceholders := "(?, ?, ?),(?, ?, ?)"
	if !strings.Contains(conn.lastQuery, expectedPlaceholders) {
		t.Errorf("unexpected placeholders: got %s", conn.lastQuery)
	}

	if len(conn.lastArgs) != 6 {
		t.Errorf("unexpected args count: got %d, want 6", len(conn.lastArgs))
	}
}

func BenchmarkBatchInsertSubscriptions(b *testing.B) {
	q, _ := newMockQueries()
	ctx := context.Background()
	params := make([]InsertSubscriptionParams, 100)
	for i := 0; i < 100; i++ {
		params[i] = InsertSubscriptionParams{UsersIdusers: 1, Pattern: "p", Method: "m"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.BatchInsertSubscriptions(ctx, params)
	}
}

func BenchmarkLoopInsertSubscriptions(b *testing.B) {
	q, _ := newMockQueries()
	ctx := context.Background()
	params := make([]InsertSubscriptionParams, 100)
	for i := 0; i < 100; i++ {
		params[i] = InsertSubscriptionParams{UsersIdusers: 1, Pattern: "p", Method: "m"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range params {
			q.InsertSubscription(ctx, p)
		}
	}
}

func BenchmarkBatchDeleteSubscriptions(b *testing.B) {
	q, _ := newMockQueries()
	ctx := context.Background()
	params := make([]DeleteSubscriptionForSubscriberParams, 100)
	for i := 0; i < 100; i++ {
		params[i] = DeleteSubscriptionForSubscriberParams{SubscriberID: 1, Pattern: "p", Method: "m"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.BatchDeleteSubscriptions(ctx, params)
	}
}

func BenchmarkLoopDeleteSubscriptions(b *testing.B) {
	q, _ := newMockQueries()
	ctx := context.Background()
	params := make([]DeleteSubscriptionForSubscriberParams, 100)
	for i := 0; i < 100; i++ {
		params[i] = DeleteSubscriptionForSubscriberParams{SubscriberID: 1, Pattern: "p", Method: "m"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range params {
			q.DeleteSubscriptionForSubscriber(ctx, p)
		}
	}
}
