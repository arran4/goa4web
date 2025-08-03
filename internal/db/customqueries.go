package db

import "context"

type CustomQueries interface {
	SearchBloggers(ctx context.Context, arg SearchBloggersParams) ([]*BloggerCountRow, error)
	ListBloggers(ctx context.Context, arg ListBloggersParams) ([]*BloggerCountRow, error)
	SearchWriters(ctx context.Context, arg SearchWritersParams) ([]*WriterCountRow, error)
	ListWriters(ctx context.Context, arg ListWritersParams) ([]*WriterCountRow, error)
	InsertFAQForWriter(ctx context.Context, arg InsertFAQForWriterParams) (int64, error)
	ListUsersFiltered(ctx context.Context, arg ListUsersFilteredParams) ([]*UserFilteredRow, error)
	SearchUsersFiltered(ctx context.Context, arg SearchUsersFilteredParams) ([]*UserFilteredRow, error)
	AdminCountForumCategories(ctx context.Context) (int64, error)
	AdminCountForumTopics(ctx context.Context) (int64, error)
	AdminCountForumThreads(ctx context.Context) (int64, error)
	AdminCountTable(ctx context.Context, table string) (int64, error)
	AdminDeleteUser(ctx context.Context, id int32) error
	AdminUpdateUserUsername(ctx context.Context, arg AdminUpdateUserUsernameParams) error
	AdminUpdateRole(ctx context.Context, arg AdminUpdateRoleParams) error
	MonthlyUsageCounts(ctx context.Context, startYear int32) ([]*MonthlyUsageRow, error)
	UserMonthlyUsageCounts(ctx context.Context, startYear int32) ([]*UserMonthlyUsageRow, error)
}
