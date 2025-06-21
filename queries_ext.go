package main

import (
	"context"
)

// GetPermissionsByUserID returns all permissions for the given user.
func (q *Queries) GetPermissionsByUserID(ctx context.Context, userID int32) ([]*Permission, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT idpermissions, users_idusers, section, level FROM permissions WHERE users_idusers = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.Idpermissions, &p.UsersIdusers, &p.Section, &p.Level); err != nil {
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
	row := q.db.QueryRowContext(ctx, "SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates FROM preferences WHERE users_idusers = ?", userID)
	var p Preference
	err := row.Scan(&p.Idpreferences, &p.LanguageIdlanguage, &p.UsersIdusers, &p.Emailforumupdates)
	return &p, err
}

// GetUserLanguages returns the language records for the user.
func (q *Queries) GetUserLanguages(ctx context.Context, userID int32) ([]*Userlang, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT iduserlang, users_idusers, language_idlanguage FROM userlang WHERE users_idusers = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Userlang
	for rows.Next() {
		var ul Userlang
		if err := rows.Scan(&ul.Iduserlang, &ul.UsersIdusers, &ul.LanguageIdlanguage); err != nil {
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
	_, err := q.db.ExecContext(ctx, "DELETE FROM userlang WHERE users_idusers = ?", userID)
	return err
}

type InsertUserLangParams struct {
	UsersIdusers       int32
	LanguageIdlanguage int32
}

// InsertUserLang adds a user language record.
func (q *Queries) InsertUserLang(ctx context.Context, arg InsertUserLangParams) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO userlang (users_idusers, language_idlanguage) VALUES (?, ?)", arg.UsersIdusers, arg.LanguageIdlanguage)
	return err
}

type InsertPreferenceParams struct {
	LanguageIdlanguage int32
	UsersIdusers       int32
}

// InsertPreference creates a new preference row for the user.
func (q *Queries) InsertPreference(ctx context.Context, arg InsertPreferenceParams) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO preferences (language_idlanguage, users_idusers) VALUES (?, ?)", arg.LanguageIdlanguage, arg.UsersIdusers)
	return err
}

type UpdatePreferenceParams struct {
	LanguageIdlanguage int32
	UsersIdusers       int32
}

// UpdatePreference updates the user's default language preference.
func (q *Queries) UpdatePreference(ctx context.Context, arg UpdatePreferenceParams) error {
	_, err := q.db.ExecContext(ctx, "UPDATE preferences SET language_idlanguage = ? WHERE users_idusers = ?", arg.LanguageIdlanguage, arg.UsersIdusers)
	return err
}

// InsertPendingEmail adds an email to the sending queue.
type InsertPendingEmailParams struct {
	ToEmail string
	Subject string
	Body    string
}

func (q *Queries) InsertPendingEmail(ctx context.Context, arg InsertPendingEmailParams) error {
	_, err := q.db.ExecContext(ctx, "INSERT INTO pending_emails (to_email, subject, body) VALUES (?, ?, ?)", arg.ToEmail, arg.Subject, arg.Body)
	return err
}

// FetchPendingEmails returns unsent queued emails up to the provided limit.
func (q *Queries) FetchPendingEmails(ctx context.Context, limit int32) ([]*PendingEmail, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT id, to_email, subject, body FROM pending_emails WHERE sent_at IS NULL ORDER BY id LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PendingEmail
	for rows.Next() {
		var p PendingEmail
		if err := rows.Scan(&p.ID, &p.ToEmail, &p.Subject, &p.Body); err != nil {
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

// ListUsers returns a limited set of users ordered by ID.
type ListUsersParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]*User, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT idusers, email, passwd, username FROM users ORDER BY idusers LIMIT ? OFFSET ?", arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Idusers, &u.Email, &u.Passwd, &u.Username); err != nil {
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
	rows, err := q.db.QueryContext(ctx, "SELECT idusers, email, passwd, username FROM users WHERE LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) ORDER BY idusers LIMIT ? OFFSET ?", like, like, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Idusers, &u.Email, &u.Passwd, &u.Username); err != nil {
			return nil, err
		}
		items = append(items, &u)
	}
	return items, rows.Err()
}

// GetPermissionsBySectionWithUsers lists permissions for a section with user info.
func (q *Queries) GetPermissionsBySectionWithUsers(ctx context.Context, section string) ([]*PermissionWithUser, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT p.idpermissions, p.users_idusers, p.section, p.level, u.username, u.email FROM permissions p JOIN users u ON u.idusers = p.users_idusers WHERE p.section = ? ORDER BY p.level", section)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PermissionWithUser
	for rows.Next() {
		var p PermissionWithUser
		if err := rows.Scan(&p.Idpermissions, &p.UsersIdusers, &p.Section, &p.Level, &p.Username, &p.Email); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, rows.Err()
}
