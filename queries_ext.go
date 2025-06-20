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
