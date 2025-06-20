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
