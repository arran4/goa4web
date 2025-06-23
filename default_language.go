package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

// validateLanguageName reports an error if name does not exist in the
// language table.
func validateLanguageName(ctx context.Context, db *sql.DB, name string) error {
	if name == "" || db == nil {
		return nil
	}
	q := &Queries{db: db}
	if _, err := q.GetLanguageIDByName(ctx, sql.NullString{String: name, Valid: true}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("language %s not found", name)
		}
		return err
	}
	return nil
}

// validateDefaultLanguage checks that cfg.DefaultLanguage exists in the language table.
func validateDefaultLanguage(ctx context.Context, db *sql.DB, cfg *RuntimeConfig) error {
	return validateLanguageName(ctx, db, cfg.DefaultLanguage)
}

// resolveDefaultLanguageID chooses the best language ID for new content.
func resolveDefaultLanguageID(r *http.Request, contextLangID int32) int {
	if contextLangID != 0 {
		return int(contextLangID)
	}
	if pref, _ := r.Context().Value(ContextValues("preference")).(*Preference); pref != nil {
		if pref.LanguageIdlanguage != 0 {
			return int(pref.LanguageIdlanguage)
		}
	}
	if appRuntimeConfig.DefaultLanguage != "" {
		queries := r.Context().Value(ContextValues("queries")).(*Queries)
		if queries != nil {
			if id, err := queries.GetLanguageIDByName(r.Context(), sql.NullString{String: appRuntimeConfig.DefaultLanguage, Valid: true}); err == nil {
				return int(id)
			}
		}
	}
	return 0
}
