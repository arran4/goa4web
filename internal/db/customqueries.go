package db

import "context"

type CustomQueries interface {
	AdminListUsersFiltered(ctx context.Context, arg AdminListUsersFilteredParams) ([]*UserFilteredRow, error)
	AdminSearchUsersFiltered(ctx context.Context, arg AdminSearchUsersFilteredParams) ([]*UserFilteredRow, error)
	MonthlyUsageCounts(ctx context.Context, startYear int32) ([]*MonthlyUsageRow, error)
	UserMonthlyUsageCounts(ctx context.Context, startYear int32) ([]*UserMonthlyUsageRow, error)
}
