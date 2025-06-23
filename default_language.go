package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"unicode"
)

// validateLanguageName ensures the provided language name contains only
// letters, digits, dashes or underscores.
func validateLanguageName(name string) error {
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_') {
			return fmt.Errorf("invalid language name")
		}
	}
	return nil
}

// validateDefaultLanguage verifies that the configured default language exists.
func validateDefaultLanguage(ctx context.Context, q *Queries, name string) error {
	if name == "" {
		return nil
	}
	if err := validateLanguageName(name); err != nil {
		return err
	}
	_, err := q.GetLanguageIDByName(ctx, sql.NullString{String: name, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unknown language %q", name)
	}
	return err
}

// resolveDefaultLanguageID converts the configured language name to its ID.
func resolveDefaultLanguageID(ctx context.Context, q *Queries) int32 {
	if appRuntimeConfig.DefaultLanguage == "" {
		return 0
	}
	id, err := q.GetLanguageIDByName(ctx, sql.NullString{String: appRuntimeConfig.DefaultLanguage, Valid: true})
	if err != nil {
		return 0
	}
	return id
}
