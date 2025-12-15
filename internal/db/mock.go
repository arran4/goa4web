package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
)

var (
	_               Querier = &QuerierProxier{}
	_               driver.Driver  = &DBProxy{}
	_               driver.Conn    = &ConnProxy{}
)

type QuerierProxier struct {
	*Querier
	OverwrittenHasGrant                                func(ctx context.Context, arg HasGrantParams) (int32, error)
	OverwrittenCreatePrivateForumTopic                 func(ctx context.Context, arg CreatePrivateForumTopicParams) (int32, error)
	OverwrittenCreatePrivateForumThread                func(ctx context.Context, arg CreatePrivateForumThreadParams) (int32, error)
	OverwrittenGrantPermission                         func(ctx context.Context, arg GrantPermissionParams) error
	OverwrittenCreateComment                           func(ctx context.Context, arg CreateCommentParams) (int32, error)
	OverwrittenSubscribeTo                               func(ctx context.Context, arg SubscribeToParams) error
	OverwrittenSystemGetLogin                          func(ctx context.Context, username string) (SystemGetLoginRow, error)
	OverwrittenSystemGetUserByUsername                 func(ctx context.Context, username string) (SystemGetUserByUsernameRow, error)
	OverwrittenGetUserByID                             func(ctx context.Context, idusers int32) (GetUserByIDRow, error)
	OverwrittenGetUserByEmail                          func(ctx context.Context, email string) (GetUserByEmailRow, error)
	OverwrittenCreateUser                              func(ctx context.Context, arg CreateUserParams) (int32, error)
	OverwrittenCreateUserEmail                         func(ctx context.Context, arg CreateUserEmailParams) (int32, error)
	OverwrittenSetUserPassword                         func(ctx context.Context, arg SetUserPasswordParams) error
	OverwrittenGetUserPermission                       func(ctx context.Context, arg GetUserPermissionParams) (int32, error)
	OverwrittenListNotificationsByUserID               func(ctx context.Context, arg ListNotificationsByUserIDParams) ([]ListNotificationsByUserIDRow, error)
	OverwrittenGetNotificationByID                     func(ctx context.Context, idnotifications int32) (GetNotificationByIDRow, error)
	OverwrittenDeleteNotification                      func(ctx context.Context, idnotifications int32) error
	OverwrittenGetUserUnreadNotificationCountByUserID  func(ctx context.Context, usersIdusers int32) (int64, error)
	OverwrittenCreateNotification                      func(ctx context.Context, arg CreateNotificationParams) error
	OverwrittenGetSubscriptionsForItem                 func(ctx context.Context, arg GetSubscriptionsForItemParams) ([]GetSubscriptionsForItemRow, error)
	OverwrittenGetEmailByID                            func(ctx context.Context, iduserEmails int32) (GetEmailByIDRow, error)
	OverwrittenGetUserEmailsByUserID                   func(ctx context.Context, usersIdusers int32) ([]GetUserEmailsByUserIDRow, error)
	OverwrittenSetUserEmailVerified                    func(ctx context.Context, iduserEmails int32) error
	OverwrittenGetLatestSchemaVersion                  func(ctx context.Context) (sql.NullInt32, error)
	OverwrittenUpdateUserPassword                      func(ctx context.Context, arg UpdateUserPasswordParams) error
}

func (q *QuerierProxier) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	if q.OverwrittenUpdateUserPassword == nil {
		panic("UpdateUserPassword not implemented")
	}
	return q.OverwrittenUpdateUserPassword(ctx, arg)
}

func (q *QuerierProxier) GetLatestSchemaVersion(ctx context.Context) (sql.NullInt32, error) {
	if q.OverwrittenGetLatestSchemaVersion == nil {
		panic("GetLatestSchemaVersion not implemented")
	}
	return q.OverwrittenGetLatestSchemaVersion(ctx)
}

func (q *QuerierProxier) GetUserEmailsByUserID(ctx context.Context, usersIdusers int32) ([]GetUserEmailsByUserIDRow, error) {
	if q.OverwrittenGetUserEmailsByUserID == nil {
		panic("GetUserEmailsByUserID not implemented")
	}
	return q.OverwrittenGetUserEmailsByUserID(ctx, usersIdusers)
}

func (q *QuerierProxier) SetUserEmailVerified(ctx context.Context, iduserEmails int32) error {
	if q.OverwrittenSetUserEmailVerified == nil {
		panic("SetUserEmailVerified not implemented")
	}
	return q.OverwrittenSetUserEmailVerified(ctx, iduserEmails)
}

func (q *QuerierProxier) GetEmailByID(ctx context.Context, iduserEmails int32) (GetEmailByIDRow, error) {
	if q.OverwrittenGetEmailByID == nil {
		panic("GetEmailByID not implemented")
	}
	return q.OverwrittenGetEmailByID(ctx, iduserEmails)
}

func (q *QuerierProxier) GetSubscriptionsForItem(ctx context.Context, arg GetSubscriptionsForItemParams) ([]GetSubscriptionsForItemRow, error) {
	if q.OverwrittenGetSubscriptionsForItem == nil {
		panic("GetSubscriptionsForItem not implemented")
	}
	return q.OverwrittenGetSubscriptionsForItem(ctx, arg)
}

func (q *QuerierProxier) CreateNotification(ctx context.Context, arg CreateNotificationParams) error {
	if q.OverwrittenCreateNotification == nil {
		panic("CreateNotification not implemented")
	}
	return q.OverwrittenCreateNotification(ctx, arg)
}

func (q *QuerierProxier) GetUserUnreadNotificationCountByUserID(ctx context.Context, usersIdusers int32) (int64, error) {
	if q.OverwrittenGetUserUnreadNotificationCountByUserID == nil {
		panic("GetUserUnreadNotificationCountByUserID not implemented")
	}
	return q.OverwrittenGetUserUnreadNotificationCountByUserID(ctx, usersIdusers)
}

func (q *QuerierProxier) DeleteNotification(ctx context.Context, idnotifications int32) error {
	if q.OverwrittenDeleteNotification == nil {
		panic("DeleteNotification not implemented")
	}
	return q.OverwrittenDeleteNotification(ctx, idnotifications)
}

func (q *QuerierProxier) GetNotificationByID(ctx context.Context, idnotifications int32) (GetNotificationByIDRow, error) {
	if q.OverwrittenGetNotificationByID == nil {
		panic("GetNotificationByID not implemented")
	}
	return q.OverwrittenGetNotificationByID(ctx, idnotifications)
}

func (q *QuerierProxier) ListNotificationsByUserID(ctx context.Context, arg ListNotificationsByUserIDParams) ([]ListNotificationsByUserIDRow, error) {
	if q.OverwrittenListNotificationsByUserID == nil {
		panic("ListNotificationsByUserID not implemented")
	}
	return q.OverwrittenListNotificationsByUserID(ctx, arg)
}

func (q *QuerierProxier) GetUserPermission(ctx context.Context, arg GetUserPermissionParams) (int32, error) {
	if q.OverwrittenGetUserPermission == nil {
		panic("GetUserPermission not implemented")
	}
	return q.OverwrittenGetUserPermission(ctx, arg)
}

func (q *QuerierProxier) SetUserPassword(ctx context.Context, arg SetUserPasswordParams) error {
	if q.OverwrittenSetUserPassword == nil {
		panic("SetUserPassword not implemented")
	}
	return q.OverwrittenSetUserPassword(ctx, arg)
}

func (q *QuerierProxier) CreateUserEmail(ctx context.Context, arg CreateUserEmailParams) (int32, error) {
	if q.OverwrittenCreateUserEmail == nil {
		panic("CreateUserEmail not implemented")
	}
	return q.OverwrittenCreateUserEmail(ctx, arg)
}

func (q *QuerierProxier) CreateUser(ctx context.Context, arg CreateUserParams) (int32, error) {
	if q.OverwrittenCreateUser == nil {
		panic("CreateUser not implemented")
	}
	return q.OverwrittenCreateUser(ctx, arg)
}

func (q *QuerierProxier) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	if q.OverwrittenGetUserByEmail == nil {
		panic("GetUserByEmail not implemented")
	}
	return q.OverwrittenGetUserByEmail(ctx, email)
}

func (q *QuerierProxier) GetUserByID(ctx context.Context, idusers int32) (GetUserByIDRow, error) {
	if q.OverwrittenGetUserByID == nil {
		panic("GetUserByID not implemented")
	}
	return q.OverwrittenGetUserByID(ctx, idusers)
}

func (q *QuerierProxier) SystemGetUserByUsername(ctx context.Context, username string) (SystemGetUserByUsernameRow, error) {
	if q.OverwrittenSystemGetUserByUsername == nil {
		panic("SystemGetUserByUsername not implemented")
	}
	return q.OverwrittenSystemGetUserByUsername(ctx, username)
}

func (q *QuerierProxier) SystemGetLogin(ctx context.Context, username string) (SystemGetLoginRow, error) {
	if q.OverwrittenSystemGetLogin == nil {
		panic("SystemGetLogin not implemented")
	}
	return q.OverwrittenSystemGetLogin(ctx, username)
}

func (q *QuerierProxier) SubscribeTo(ctx context.Context, arg SubscribeToParams) error {
	if q.OverwrittenSubscribeTo == nil {
		panic("SubscribeTo not implemented")
	}
	return q.OverwrittenSubscribeTo(ctx, arg)
}

func (q *QuerierProxier) CreateComment(ctx context.Context, arg CreateCommentParams) (int32, error) {
	if q.OverwrittenCreateComment == nil {
		panic("CreateComment not implemented")
	}
	return q.OverwrittenCreateComment(ctx, arg)
}

func (q *QuerierProxier) GrantPermission(ctx context.Context, arg GrantPermissionParams) error {
	if q.OverwrittenGrantPermission == nil {
		panic("GrantPermission not implemented")
	}
	return q.OverwrittenGrantPermission(ctx, arg)
}

func (q *QuerierProxier) CreatePrivateForumThread(ctx context.Context, arg CreatePrivateForumThreadParams) (int32, error) {
	if q.OverwrittenCreatePrivateForumThread == nil {
		panic("CreatePrivateForumThread not implemented")
	}
	return q.OverwrittenCreatePrivateForumThread(ctx, arg)
}

func (q *QuerierProxier) CreatePrivateForumTopic(ctx context.Context, arg CreatePrivateForumTopicParams) (int32, error) {
	if q.OverwrittenCreatePrivateForumTopic == nil {
		panic("CreatePrivateForumTopic not implemented")
	}
	return q.OverwrittenCreatePrivateForumTopic(ctx, arg)
}

func (q *QuerierProxier) HasGrant(ctx context.Context, arg HasGrantParams) (int32, error) {
	if q.OverwrittenHasGrant == nil {
		panic("HasGrant not implemented")
	}
	return q.OverwrittenHasGrant(ctx, arg)
}

type DBProxy struct {
	driver.Driver
	OverwrittenOpen func(name string) (driver.Conn, error)
}

func (d *DBProxy) Open(name string) (driver.Conn, error) {
	if d.OverwrittenOpen == nil {
		panic("Open not implemented")
	}
	return d.OverwrittenOpen(name)
}

type ConnProxy struct {
	driver.Conn
	OverwrittenBegin  func() (driver.Tx, error)
	OverwrittenQueryContext func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error)
	OverwrittenExecContext func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error)
}

func (c *ConnProxy) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.OverwrittenQueryContext == nil {
		panic("QueryContext not implemented")
	}
	return c.OverwrittenQueryContext(ctx, query, args)
}

func (c *ConnProxy) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.OverwrittenExecContext == nil {
		panic("ExecContext not implemented")
	}
	return c.OverwrittenExecContext(ctx, query, args)
}

func (c *ConnProxy) Prepare(query string) (driver.Stmt, error) {
	return &StmtProxy{
		query: query,
		conn:  c,
	}, nil
}

func (c *ConnProxy) Close() error {
	return nil
}

func (c *ConnProxy) Begin() (driver.Tx, error) {
	if c.OverwrittenBegin == nil {
		panic("Begin not implemented")
	}
	return c.OverwrittenBegin()
}

type StmtProxy struct {
	driver.Stmt
	query string
	conn  *ConnProxy
}

func (s *StmtProxy) Close() error {
	return nil
}

func (s *StmtProxy) NumInput() int {
	return -1
}

func (s *StmtProxy) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.ExecContext(context.Background(), s.query, valuesToNamedValues(args))
}

func (s *StmtProxy) Query(args []driver.Value) (driver.Rows, error) {
	return s.conn.QueryContext(context.Background(), s.query, valuesToNamedValues(args))
}

type TxProxy struct {
	driver.Tx
	OverwrittenCommit   func() error
	OverwrittenRollback func() error
}

func (t *TxProxy) Commit() error {
	if t.OverwrittenCommit == nil {
		panic("Commit not implemented")
	}
	return t.OverwrittenCommit()
}

func (t *TxProxy) Rollback() error {
	if t.OverwrittenRollback == nil {
		panic("Rollback not implemented")
	}
	return t.OverwrittenRollback()
}

func valuesToNamedValues(values []driver.Value) []driver.NamedValue {
	namedValues := make([]driver.NamedValue, len(values))
	for i, value := range values {
		namedValues[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   value,
		}
	}
	return namedValues
}

type MockRows struct {
	driver.Rows
	Columns     []string
	Rows        [][]driver.Value
	CurrentRow  int
	CloseError  error
	closed      bool
}

func NewMockRows(columns ...string) *MockRows {
	return &MockRows{
		Columns: columns,
	}
}

func (m *MockRows) AddRow(values ...driver.Value) *MockRows {
	if len(values) != len(m.Columns) {
		panic(fmt.Sprintf("number of values %d does not match number of columns %d", len(values), len(m.Columns)))
	}
	m.Rows = append(m.Rows, values)
	return m
}

func (m *MockRows) Columns() []string {
	return m.Columns
}

func (m *MockRows) Close() error {
	m.closed = true
	return m.CloseError
}

func (m *MockRows) Next(dest []driver.Value) error {
	if m.closed {
		return io.EOF
	}
	if m.CurrentRow >= len(m.Rows) {
		return io.EOF
	}
	for i, val := range m.Rows[m.CurrentRow] {
		dest[i] = val
	}
	m.CurrentRow++
	return nil
}

func (m *MockRows) ColumnTypeDatabaseTypeName(index int) string {
	return "TEXT"
}

func (m_ *MockRows) ColumnTypeNullable(index int) (nullable, ok bool) {
	return true, true
}

func (m_ *MockRows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return 0, 0, false
}

func (m *MockRows) ColumnTypeScanType(index int) reflect.Type {
	return reflect.TypeOf("")
}

func (m *MockRows) ColumnTypeLength(index int) (length int64, ok bool) {
	return 0, false
}
