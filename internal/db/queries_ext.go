package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

// GetPermissionsByUserID returns all roles for the given user.
func (q *Queries) GetPermissionsByUserID(ctx context.Context, userID int32) ([]*GetUserRolesRow, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT ur.idpermissions, ur.users_idusers, r.name FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetUserRolesRow
	for rows.Next() {
		var p GetUserRolesRow
		if err := rows.Scan(&p.Idpermissions, &p.UsersIdusers, &p.Role); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetPreferenceByUserID returns the preference row for the user.
func (q *Queries) GetPreferenceByUserID(ctx context.Context, userID int32) (*Preference, error) {
	row := q.db.QueryRowContext(ctx, "SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates, page_size, auto_subscribe_replies FROM preferences WHERE users_idusers = ?", userID)
	var p Preference
	err := row.Scan(&p.Idpreferences, &p.LanguageIdlanguage, &p.UsersIdusers, &p.Emailforumupdates, &p.PageSize, &p.AutoSubscribeReplies)
	return &p, err
}

// GetUserLanguages returns the language records for the user.
func (q *Queries) GetUserLanguages(ctx context.Context, userID int32) ([]*UserLanguage, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT iduser_language, users_idusers, language_idlanguage FROM user_language WHERE users_idusers = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*UserLanguage
	for rows.Next() {
		var ul UserLanguage
		if err := rows.Scan(&ul.IduserLanguage, &ul.UsersIdusers, &ul.LanguageIdlanguage); err != nil {
			return nil, err
		}
		items = append(items, &ul)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteUserLanguagesByUser removes all language selections for a user.
func (q *Queries) DeleteUserLanguagesByUser(ctx context.Context, userID int32) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM user_language WHERE users_idusers = ?", userID)
	return err
}

type InsertUserLangParams struct {
	UsersIdusers       int32
	LanguageIdlanguage int32
}

// InsertUserLang adds a user language record.
func (q *Queries) InsertUserLang(ctx context.Context, arg InsertUserLangParams) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO user_language (users_idusers, language_idlanguage) VALUES (?, ?)", arg.UsersIdusers, arg.LanguageIdlanguage)
	return err
}

type InsertPreferenceParams struct {
	LanguageIdlanguage int32
	UsersIdusers       int32
	PageSize           int32
}

// InsertPreference creates a new preference row for the user.
func (q *Queries) InsertPreference(ctx context.Context, arg InsertPreferenceParams) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO preferences (language_idlanguage, users_idusers, page_size) VALUES (?, ?, ?)", arg.LanguageIdlanguage, arg.UsersIdusers, arg.PageSize)
	return err
}

type UpdatePreferenceParams struct {
	LanguageIdlanguage int32
	UsersIdusers       int32
	PageSize           int32
}

// UpdatePreference updates the user's default language preference.
func (q *Queries) UpdatePreference(ctx context.Context, arg UpdatePreferenceParams) error {
	_, err := q.db.ExecContext(ctx, "UPDATE preferences SET language_idlanguage = ?, page_size=? WHERE users_idusers = ?", arg.LanguageIdlanguage, arg.PageSize, arg.UsersIdusers)
	return err
}

// InsertPendingEmail adds an email to the sending queue.
type InsertPendingEmailParams struct {
	ToUserID int32
	Body     string
}

func (q *Queries) InsertPendingEmail(ctx context.Context, arg InsertPendingEmailParams) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO pending_emails (to_user_id, body) VALUES (?, ?)", arg.ToUserID, arg.Body)
	return err
}

// FetchPendingEmails returns unsent queued emails up to the provided limit.
func (q *Queries) FetchPendingEmails(ctx context.Context, limit int32) ([]*PendingEmail, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT id, to_user_id, body, error_count FROM pending_emails WHERE sent_at IS NULL ORDER BY id LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PendingEmail
	for rows.Next() {
		var p PendingEmail
		if err := rows.Scan(&p.ID, &p.ToUserID, &p.Body, &p.ErrorCount); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, rows.Err()
}

// MarkEmailSent updates a pending email once successfully delivered.
func (q *Queries) MarkEmailSent(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, "UPDATE pending_emails SET sent_at = NOW() WHERE id = ?", id)
	return err
}

// ListUnsentPendingEmails returns all queued emails that have not been sent yet.
func (q *Queries) ListUnsentPendingEmails(ctx context.Context) ([]*PendingEmail, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT id, to_user_id, body, error_count, created_at FROM pending_emails WHERE sent_at IS NULL ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PendingEmail
	for rows.Next() {
		var p PendingEmail
		if err := rows.Scan(&p.ID, &p.ToUserID, &p.Body, &p.ErrorCount, &p.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, rows.Err()
}

// GetPendingEmailByID returns a single pending email.
func (q *Queries) GetPendingEmailByID(ctx context.Context, id int32) (*PendingEmail, error) {
	row := q.db.QueryRowContext(ctx, "SELECT id, to_user_id, body, error_count FROM pending_emails WHERE id = ?", id)
	var p PendingEmail
	err := row.Scan(&p.ID, &p.ToUserID, &p.Body, &p.ErrorCount)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeletePendingEmail removes an email from the queue.
func (q *Queries) DeletePendingEmail(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, "DELETE FROM pending_emails WHERE id = ?", id)
	return err
}

// IncrementEmailError increases the error count for a queued email and returns the updated value.
func (q *Queries) IncrementEmailError(ctx context.Context, id int32) (int32, error) {
	if _, err := q.db.ExecContext(ctx, "UPDATE pending_emails SET error_count = error_count + 1 WHERE id = ?", id); err != nil {
		return 0, err
	}
	var count int32
	err := q.db.QueryRowContext(ctx, "SELECT error_count FROM pending_emails WHERE id = ?", id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListUsers returns a limited set of users ordered by ID.
type ListUsersParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]*User, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT idusers, email, username FROM users ORDER BY idusers LIMIT ? OFFSET ?", arg.Limit, arg.Offset)
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

// SearchUsers finds users by username or email with pagination.
type SearchUsersParams struct {
	Query  string
	Limit  int32
	Offset int32
}

func (q *Queries) SearchUsers(ctx context.Context, arg SearchUsersParams) ([]*User, error) {
	like := "%" + arg.Query + "%"
	rows, err := q.db.QueryContext(ctx, "SELECT idusers, email, username FROM users WHERE LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) ORDER BY idusers LIMIT ? OFFSET ?", like, like, arg.Limit, arg.Offset)
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
		cond = append(cond, "u.status = ?")
		args = append(args, arg.Status)
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
		cond = append(cond, "u.status = ?")
		args = append(args, arg.Status)
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

// ListUserIDsByRole returns all user IDs with the specified role.
func (q *Queries) ListUserIDsByRole(ctx context.Context, role string) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT u.idusers FROM users u JOIN user_roles ur ON ur.users_idusers = u.idusers JOIN roles r ON ur.role_id = r.id WHERE r.name = ? ORDER BY u.idusers",
		role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// AllUserIDs returns all user IDs in the system.
func (q *Queries) AllUserIDs(ctx context.Context) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT idusers FROM users ORDER BY idusers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// UsersByID returns a map of user ID to User for the provided IDs.
func (q *Queries) UsersByID(ctx context.Context, ids []int32) (map[int32]*User, error) {
	if len(ids) == 0 {
		return map[int32]*User{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("SELECT idusers, username FROM users WHERE idusers IN (%s)", strings.Join(placeholders, ","))
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make(map[int32]*User)
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Idusers, &u.Username); err != nil {
			return nil, err
		}
		users[u.Idusers] = &u
	}
	return users, rows.Err()
}

// BloggerCountRow includes a username with their blog post count.
type BloggerCountRow struct {
	Username sql.NullString
	Count    int64
}

// ListBloggers returns bloggers with the number of posts, ordered by username.
type ListBloggersParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListBloggers(ctx context.Context, arg ListBloggersParams) ([]*BloggerCountRow, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT u.username, COUNT(b.idblogs) FROM blogs b JOIN users u ON b.users_idusers = u.idusers GROUP BY u.idusers ORDER BY u.username LIMIT ? OFFSET ?",
		arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*BloggerCountRow
	for rows.Next() {
		var i BloggerCountRow
		if err := rows.Scan(&i.Username, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	return items, rows.Err()
}

// SearchBloggers finds bloggers by username or email with pagination.
type SearchBloggersParams struct {
	Query  string
	Limit  int32
	Offset int32
}

func (q *Queries) SearchBloggers(ctx context.Context, arg SearchBloggersParams) ([]*BloggerCountRow, error) {
	like := "%" + arg.Query + "%"
	rows, err := q.db.QueryContext(ctx,
		"SELECT u.username, COUNT(b.idblogs) FROM blogs b JOIN users u ON b.users_idusers = u.idusers WHERE LOWER(u.username) LIKE LOWER(?) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(?) GROUP BY u.idusers ORDER BY u.username LIMIT ? OFFSET ?",
		like, like, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*BloggerCountRow
	for rows.Next() {
		var i BloggerCountRow
		if err := rows.Scan(&i.Username, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	return items, rows.Err()
}

// RecentNotifications returns the newest notifications across all users limited by the provided count.
func (q *Queries) RecentNotifications(ctx context.Context, limit int32) ([]*Notification, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT id, users_idusers, link, message, created_at, read_at FROM notifications ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UsersIdusers, &n.Link, &n.Message, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, &n)
	}
	return items, rows.Err()
}

// CountThreadsByBoard returns the number of unique threads for a board.
func (q *Queries) CountThreadsByBoard(ctx context.Context, boardID int32) (int32, error) {
	var c int32
	err := q.db.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT forumthread_id) FROM imagepost WHERE imageboard_idimageboard = ?",
		boardID).Scan(&c)
	return c, err
}

// BoardPostCountRow represents a board or topic name with a thread count.
type BoardPostCountRow struct {
	Name  sql.NullString
	Count int64
}

// CategoryCountRow represents a category name with a content count.
// It mirrors BoardPostCountRow for reuse in different contexts.
type CategoryCountRow = BoardPostCountRow

// ForumTopicThreadCounts returns the number of threads per forum topic ordered by title.
func (q *Queries) ForumTopicThreadCounts(ctx context.Context) ([]*BoardPostCountRow, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT t.title, COUNT(th.idforumthread) FROM forumtopic t LEFT JOIN forumthread th ON th.forumtopic_idforumtopic = t.idforumtopic GROUP BY t.idforumtopic ORDER BY t.title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*BoardPostCountRow
	for rows.Next() {
		var i BoardPostCountRow
		if err := rows.Scan(&i.Name, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	return items, rows.Err()
}

// ForumCategoryThreadCounts returns the number of threads per forum category ordered by title.
func (q *Queries) ForumCategoryThreadCounts(ctx context.Context) ([]*CategoryCountRow, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT c.title, COUNT(th.idforumthread) FROM forumcategory c "+
			"LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory "+
			"LEFT JOIN forumthread th ON th.forumtopic_idforumtopic = t.idforumtopic "+
			"GROUP BY c.idforumcategory ORDER BY c.title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CategoryCountRow
	for rows.Next() {
		var i CategoryCountRow
		if err := rows.Scan(&i.Name, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	return items, rows.Err()
}

// ImageboardPostCounts returns the number of posts per image board ordered by title.
func (q *Queries) ImageboardPostCounts(ctx context.Context) ([]*BoardPostCountRow, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT ib.title, COUNT(ip.idimagepost) FROM imageboard ib LEFT JOIN imagepost ip ON ip.imageboard_idimageboard = ib.idimageboard GROUP BY ib.idimageboard ORDER BY ib.title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*BoardPostCountRow
	for rows.Next() {
		var i BoardPostCountRow
		if err := rows.Scan(&i.Name, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	return items, rows.Err()
}

// WritingCategoryCounts returns the number of writings per writing category ordered by title.
func (q *Queries) WritingCategoryCounts(ctx context.Context) ([]*CategoryCountRow, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT wc.title, COUNT(w.idwriting) FROM writing_category wc "+
			"LEFT JOIN writing w ON w.writing_category_id = wc.idwritingCategory "+
			"GROUP BY wc.idwritingCategory ORDER BY wc.title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CategoryCountRow
	for rows.Next() {
		var i CategoryCountRow
		if err := rows.Scan(&i.Name, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	return items, rows.Err()
}

// UserPostCountRow aggregates content counts for a user.
type UserPostCountRow struct {
	Username sql.NullString
	Blogs    int64
	News     int64
	Comments int64
	Images   int64
	Links    int64
	Writings int64
}

// UserPostCounts returns aggregated post counts for each user.
func (q *Queries) UserPostCounts(ctx context.Context) ([]*UserPostCountRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT u.username,
        COUNT(DISTINCT b.idblogs) AS blogs,
        COUNT(DISTINCT n.idsiteNews) AS news,
        COUNT(DISTINCT c.idcomments) AS comments,
        COUNT(DISTINCT i.idimagepost) AS images,
        COUNT(DISTINCT l.idlinker) AS links,
        COUNT(DISTINCT w.idwriting) AS writings
        FROM users u
        LEFT JOIN blogs b ON b.users_idusers = u.idusers
        LEFT JOIN siteNews n ON n.users_idusers = u.idusers
        LEFT JOIN comments c ON c.users_idusers = u.idusers
        LEFT JOIN imagepost i ON i.users_idusers = u.idusers
        LEFT JOIN linker l ON l.users_idusers = u.idusers
        LEFT JOIN writing w ON w.users_idusers = u.idusers
        GROUP BY u.idusers
        ORDER BY u.username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*UserPostCountRow
	for rows.Next() {
		var i UserPostCountRow
		if err := rows.Scan(&i.Username, &i.Blogs, &i.News, &i.Comments, &i.Images, &i.Links, &i.Writings); err != nil {
			return nil, err
		}
		items = append(items, &i)
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

// monthlyCounts is a helper that returns post counts grouped by year and month.
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

// userMonthlyCounts returns post counts grouped by user, year and month.
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

// MonthlyUsageCounts merges monthly counts from several content tables.
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

// UserMonthlyUsageCounts merges monthly usage counts for each user across several tables.
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

// GetTemplateOverride retrieves the stored override for the named template.
func (q *Queries) GetTemplateOverride(ctx context.Context, name string) (string, error) {
	row := q.db.QueryRowContext(ctx, "SELECT body FROM template_overrides WHERE name = ?", name)
	var body string
	switch err := row.Scan(&body); err {
	case sql.ErrNoRows:
		return "", nil
	default:
		return body, err
	}
}

// SetTemplateOverride stores or updates a template override.
func (q *Queries) SetTemplateOverride(ctx context.Context, name, body string) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO template_overrides (name, body) VALUES (?, ?) ON DUPLICATE KEY UPDATE body = VALUES(body)", name, body)
	return err
}

// UserInfoRow represents basic user details with optional admin status and
// account creation time.
type UserInfoRow struct {
	ID        int32
	Username  sql.NullString
	Email     sql.NullString
	Admin     bool
	CreatedAt sql.NullTime
}

// ListUserInfo returns all users along with administrator flag and the earliest
// session creation time if available.
func (q *Queries) ListUserInfo(ctx context.Context) ([]*UserInfoRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT u.idusers, u.username,
                (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email,
                IF(r.id IS NULL, 0, 1) AS admin,
                MIN(s.created_at) AS created_at
                FROM users u
               LEFT JOIN user_roles ur ON ur.users_idusers = u.idusers
               LEFT JOIN roles r ON ur.role_id = r.id AND r.name = 'administrator'
               LEFT JOIN sessions s ON s.users_idusers = u.idusers
               GROUP BY u.idusers
               ORDER BY u.idusers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*UserInfoRow
	for rows.Next() {
		var u UserInfoRow
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Admin, &u.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// RecentAuditLogRow represents an audit log entry with optional username.
type RecentAuditLogRow struct {
	ID           int32
	UsersIdusers int32
	Username     sql.NullString
	Action       string
	CreatedAt    time.Time
}

// GetRecentAuditLogs returns the newest audit log entries limited by the count.
func (q *Queries) GetRecentAuditLogs(ctx context.Context, limit int32) ([]*RecentAuditLogRow, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT a.id, a.users_idusers, u.username, a.action, a.created_at
        FROM audit_log a LEFT JOIN users u ON a.users_idusers = u.idusers
        ORDER BY a.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*RecentAuditLogRow
	for rows.Next() {
		var r RecentAuditLogRow
		if err := rows.Scan(&r.ID, &r.UsersIdusers, &r.Username, &r.Action, &r.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &r)
	}
	return items, rows.Err()
}
