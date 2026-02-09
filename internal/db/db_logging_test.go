package db

import (
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConnector struct {
	conn driver.Conn
	err  error
}

func (m *mockConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return m.conn, m.err
}

func (m *mockConnector) Driver() driver.Driver {
	return nil
}

type mockConn struct {
}

func (m *mockConn) Prepare(query string) (driver.Stmt, error) {
	return nil, nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) Begin() (driver.Tx, error) {
	return nil, nil
}

func TestNewLoggingConnector(t *testing.T) {
	t.Run("verbosity 0 returns base", func(t *testing.T) {
		base := &mockConnector{}
		got := NewLoggingConnector(base, 0)
		assert.Equal(t, base, got)
	})

	t.Run("verbosity > 0 returns wrapped", func(t *testing.T) {
		base := &mockConnector{}
		got := NewLoggingConnector(base, 1)
		_, ok := got.(loggingConnector)
		assert.True(t, ok)
	})
}

func TestLoggingConnector_Connect(t *testing.T) {
	// Need to capture log output, so we can't run in parallel with tests that depend on log output
	originalOutput := log.Writer()
	defer log.SetOutput(originalOutput)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	t.Run("logs connection success", func(t *testing.T) {
		buf.Reset()
		base := &mockConnector{conn: &mockConn{}}
		lc := NewLoggingConnector(base, 1)

		conn, err := lc.Connect(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		assert.Contains(t, buf.String(), "db connection")
		assert.Contains(t, buf.String(), "opened")

		// Verify loggingConn is returned
		_, ok := conn.(*loggingConn)
		assert.True(t, ok, "expected loggingConn type")
	})

	t.Run("logs connection error", func(t *testing.T) {
		buf.Reset()
		expectedErr := errors.New("connection failed")
		base := &mockConnector{err: expectedErr}
		lc := NewLoggingConnector(base, 1)

		conn, err := lc.Connect(context.Background())
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Equal(t, expectedErr, err)
		assert.Contains(t, buf.String(), "DB connect error: connection failed")
	})
}
