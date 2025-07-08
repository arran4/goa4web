package language

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"unicode"

	db "github.com/arran4/goa4web/internal/db"
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

// ValidateDefaultLanguage verifies that the configured default language exists.
func ValidateDefaultLanguage(ctx context.Context, q *db.Queries, name string) error {
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

// ResolveDefaultLanguageID converts the configured language name to its ID.
func ResolveDefaultLanguageID(ctx context.Context, q *db.Queries, name string) int32 {
	if name == "" {
		return 0
	}
	id, err := q.GetLanguageIDByName(ctx, sql.NullString{String: name, Valid: true})
	if err != nil {
		return 0
	}
	return id
}
