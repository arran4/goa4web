package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// BloggerCountRow includes a username with the number of posts.
type BloggerCountRow struct {
	Username sql.NullString
	Count    int64
}

// ListBloggers returns bloggers with the number of posts, ordered by username.
type ListBloggersParams struct {
	ViewerID int32
	Limit    int32
	Offset   int32
}

func (q *Queries) ListBloggers(ctx context.Context, arg ListBloggersParams) ([]*BloggerCountRow, error) {
	rows, err := q.ListBloggersForViewer(ctx, ListBloggersForViewerParams{
		ViewerID: arg.ViewerID,
		UserID:   sql.NullInt32{Int32: arg.ViewerID, Valid: arg.ViewerID != 0},
		Limit:    arg.Limit,
		Offset:   arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*BloggerCountRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, &BloggerCountRow{Username: r.Username, Count: r.Count})
	}
	return items, nil
}

// SearchBloggers finds bloggers by username or email with pagination.
type SearchBloggersParams struct {
	ViewerID int32
	Query    string
	Limit    int32
	Offset   int32
}

func (q *Queries) SearchBloggers(ctx context.Context, arg SearchBloggersParams) ([]*BloggerCountRow, error) {
	like := "%" + arg.Query + "%"
	rows, err := q.SearchBloggersForViewer(ctx, SearchBloggersForViewerParams{
		ViewerID: arg.ViewerID,
		Query:    like,
		UserID:   sql.NullInt32{Int32: arg.ViewerID, Valid: arg.ViewerID != 0},
		Limit:    arg.Limit,
		Offset:   arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*BloggerCountRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, &BloggerCountRow{Username: r.Username, Count: r.Count})
	}
	return items, nil
}

// ListUsersFiltered returns users filtered by role and status with pagination.
type ListUsersFilteredParams struct {
	Role   string
	Status string
	Limit  int32
	Offset int32
}

func (q *Queries) ListUsersFiltered(ctx context.Context, arg ListUsersFilteredParams) ([]*User, error) {
	query := "SELECT u.idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, u.username FROM users u"
	var args []interface{}
	var cond []string
	if arg.Role != "" {
		query += " JOIN user_roles ur ON ur.users_idusers = u.idusers JOIN roles r ON ur.role_id = r.id"
		cond = append(cond, "r.name = ?")
		args = append(args, arg.Role)
	}
	if arg.Status != "" {
		switch arg.Status {
		case "pending":
			cond = append(cond, "NOT EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = u.idusers AND r.name IN ('user','rejected'))")
		case "active":
			cond = append(cond, "EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = u.idusers AND r.name = 'user')")
		case "rejected":
			cond = append(cond, "EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = u.idusers AND r.name = 'rejected')")
		}
	}
	if len(cond) > 0 {
		query += " WHERE " + strings.Join(cond, " AND ")
	}
	query += " ORDER BY u.idusers LIMIT ? OFFSET ?"
	args = append(args, arg.Limit, arg.Offset)
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var u User
		var email sql.NullString
		if err := rows.Scan(&u.Idusers, &email, &u.Username); err != nil {
			return nil, err
		}
		items = append(items, &u)
	}
	return items, rows.Err()
}

// SearchUsersFiltered finds users by username or email with role and status filters.
type SearchUsersFilteredParams struct {
	Query  string
	Role   string
	Status string
	Limit  int32
	Offset int32
}

func (q *Queries) SearchUsersFiltered(ctx context.Context, arg SearchUsersFilteredParams) ([]*User, error) {
	like := "%" + arg.Query + "%"
	query := "SELECT u.idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, u.username FROM users u"
	var args []interface{}
	var cond []string
	if arg.Role != "" {
		query += " JOIN user_roles ur ON ur.users_idusers = u.idusers JOIN roles r ON ur.role_id = r.id"
		cond = append(cond, "r.name = ?")
		args = append(args, arg.Role)
	}
	cond = append(cond, "(LOWER(u.username) LIKE LOWER(?) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(?))")
	args = append(args, like, like)
	if arg.Status != "" {
		switch arg.Status {
		case "pending":
			cond = append(cond, "NOT EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = u.idusers AND r.name IN ('user','rejected'))")
		case "active":
			cond = append(cond, "EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = u.idusers AND r.name = 'user')")
		case "rejected":
			cond = append(cond, "EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = u.idusers AND r.name = 'rejected')")
		}
	}
	query += " WHERE " + strings.Join(cond, " AND ")
	query += " ORDER BY u.idusers LIMIT ? OFFSET ?"
	args = append(args, arg.Limit, arg.Offset)
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var u User
		var email sql.NullString
		if err := rows.Scan(&u.Idusers, &email, &u.Username); err != nil {
			return nil, err
		}
		items = append(items, &u)
	}
	return items, rows.Err()
}

// MonthlyUsageRow aggregates monthly post counts across the site.
type MonthlyUsageRow struct {
	Year     int32
	Month    int32
	Blogs    int64
	News     int64
	Comments int64
	Images   int64
	Links    int64
}

// UserMonthlyUsageRow aggregates monthly post counts for a single user.
type UserMonthlyUsageRow struct {
	Username sql.NullString
	Year     int32
	Month    int32
	Blogs    int64
	News     int64
	Comments int64
	Images   int64
	Links    int64
	Writings int64
}

func (q *Queries) monthlyCounts(ctx context.Context, table, column string, startYear int32) (map[[2]int32]int64, error) {
	query := fmt.Sprintf("SELECT YEAR(%s), MONTH(%s), COUNT(*) FROM %s WHERE YEAR(%s) >= ? GROUP BY YEAR(%s), MONTH(%s)", column, column, table, column, column, column)
	rows, err := q.db.QueryContext(ctx, query, startYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[[2]int32]int64)
	for rows.Next() {
		var year, month int32
		var count int64
		if err := rows.Scan(&year, &month, &count); err != nil {
			return nil, err
		}
		m[[2]int32{year, month}] = count
	}
	return m, rows.Err()
}

func (q *Queries) userMonthlyCounts(ctx context.Context, table, column string, startYear int32) (map[string]map[[2]int32]int64, error) {
	query := fmt.Sprintf("SELECT u.username, YEAR(%s), MONTH(%s), COUNT(*) FROM %s t JOIN users u ON t.users_idusers = u.idusers WHERE YEAR(%s) >= ? GROUP BY u.idusers, YEAR(%s), MONTH(%s)", column, column, table, column, column, column)
	rows, err := q.db.QueryContext(ctx, query, startYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make(map[string]map[[2]int32]int64)
	for rows.Next() {
		var user string
		var year, month int32
		var count int64
		if err := rows.Scan(&user, &year, &month, &count); err != nil {
			return nil, err
		}
		m, ok := data[user]
		if !ok {
			m = make(map[[2]int32]int64)
			data[user] = m
		}
		m[[2]int32{year, month}] = count
	}
	return data, rows.Err()
}

func (q *Queries) MonthlyUsageCounts(ctx context.Context, startYear int32) ([]*MonthlyUsageRow, error) {
	types := []struct {
		table  string
		column string
		set    func(*MonthlyUsageRow, int64)
	}{
		{"blogs", "written", func(r *MonthlyUsageRow, n int64) { r.Blogs = n }},
		{"siteNews", "occurred", func(r *MonthlyUsageRow, n int64) { r.News = n }},
		{"comments", "written", func(r *MonthlyUsageRow, n int64) { r.Comments = n }},
		{"imagepost", "posted", func(r *MonthlyUsageRow, n int64) { r.Images = n }},
		{"linker", "listed", func(r *MonthlyUsageRow, n int64) { r.Links = n }},
	}

	data := make(map[[2]int32]*MonthlyUsageRow)
	for _, t := range types {
		counts, err := q.monthlyCounts(ctx, t.table, t.column, startYear)
		if err != nil {
			return nil, err
		}
		for ym, c := range counts {
			row, ok := data[ym]
			if !ok {
				row = &MonthlyUsageRow{Year: ym[0], Month: ym[1]}
				data[ym] = row
			}
			t.set(row, c)
		}
	}

	keys := make([][2]int32, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] == keys[j][0] {
			return keys[i][1] < keys[j][1]
		}
		return keys[i][0] < keys[j][0]
	})

	var rows []*MonthlyUsageRow
	for _, k := range keys {
		rows = append(rows, data[k])
	}
	return rows, nil
}

func (q *Queries) UserMonthlyUsageCounts(ctx context.Context, startYear int32) ([]*UserMonthlyUsageRow, error) {
	types := []struct {
		table  string
		column string
		set    func(*UserMonthlyUsageRow, int64)
	}{
		{"blogs", "written", func(r *UserMonthlyUsageRow, n int64) { r.Blogs = n }},
		{"siteNews", "occurred", func(r *UserMonthlyUsageRow, n int64) { r.News = n }},
		{"comments", "written", func(r *UserMonthlyUsageRow, n int64) { r.Comments = n }},
		{"imagepost", "posted", func(r *UserMonthlyUsageRow, n int64) { r.Images = n }},
		{"linker", "listed", func(r *UserMonthlyUsageRow, n int64) { r.Links = n }},
		{"writing", "published", func(r *UserMonthlyUsageRow, n int64) { r.Writings = n }},
	}

	data := make(map[string]map[[2]int32]*UserMonthlyUsageRow)
	for _, t := range types {
		counts, err := q.userMonthlyCounts(ctx, t.table, t.column, startYear)
		if err != nil {
			return nil, err
		}
		for user, months := range counts {
			m, ok := data[user]
			if !ok {
				m = make(map[[2]int32]*UserMonthlyUsageRow)
				data[user] = m
			}
			for ym, c := range months {
				row, ok := m[ym]
				if !ok {
					row = &UserMonthlyUsageRow{Username: sql.NullString{String: user, Valid: true}, Year: ym[0], Month: ym[1]}
					m[ym] = row
				}
				t.set(row, c)
			}
		}
	}

	var keys []struct {
		user string
		ym   [2]int32
	}
	for user, months := range data {
		for ym := range months {
			keys = append(keys, struct {
				user string
				ym   [2]int32
			}{user, ym})
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].user == keys[j].user {
			if keys[i].ym[0] == keys[j].ym[0] {
				return keys[i].ym[1] < keys[j].ym[1]
			}
			return keys[i].ym[0] < keys[j].ym[0]
		}
		return keys[i].user < keys[j].user
	})

	var rows []*UserMonthlyUsageRow
	for _, k := range keys {
		rows = append(rows, data[k.user][k.ym])
	}
	return rows, nil
}

// WriterCountRow includes a username with the number of writings.
type WriterCountRow struct {
	Username sql.NullString
	Count    int64
}

// ListWriters returns writers with the number of writings, ordered by username.
// These wrappers provide a stable API around the generated sqlc queries.
type ListWritersParams struct {
	ViewerID int32
	Limit    int32
	Offset   int32
}

func (q *Queries) ListWriters(ctx context.Context, arg ListWritersParams) ([]*WriterCountRow, error) {
	rows, err := q.ListWritersForViewer(ctx, ListWritersForViewerParams{
		ViewerID: arg.ViewerID,
		UserID:   sql.NullInt32{Int32: arg.ViewerID, Valid: arg.ViewerID != 0},
		Limit:    arg.Limit,
		Offset:   arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*WriterCountRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, &WriterCountRow{Username: r.Username, Count: r.Count})
	}
	return items, nil
}

// SearchWriters finds writers by username or email with pagination.
// The query string is wrapped in % to match partial names.
type SearchWritersParams struct {
	ViewerID int32
	Query    string
	Limit    int32
	Offset   int32
}

func (q *Queries) SearchWriters(ctx context.Context, arg SearchWritersParams) ([]*WriterCountRow, error) {
	like := "%" + arg.Query + "%"
	rows, err := q.SearchWritersForViewer(ctx, SearchWritersForViewerParams{
		ViewerID: arg.ViewerID,
		Query:    like,
		UserID:   sql.NullInt32{Int32: arg.ViewerID, Valid: arg.ViewerID != 0},
		Limit:    arg.Limit,
		Offset:   arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*WriterCountRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, &WriterCountRow{Username: r.Username, Count: r.Count})
	}
	return items, nil
}
