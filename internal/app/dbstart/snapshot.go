package dbstart

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/migrations"
	"github.com/pressly/goose/v3"
)

// RecordMySQLSnapshotBaseline replaces version markers imported with a fresh
// canonical snapshot. Existing installations must use the migration path.
func RecordMySQLSnapshotBaseline(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(migrations.ForDriver("mysql"), ".")
	if err != nil {
		return fmt.Errorf("read embedded MySQL migrations: %w", err)
	}
	// Goose uses version zero as the starting marker.
	versions := []int64{0}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, err := goose.NumericComponent(entry.Name())
		if err != nil {
			return err
		}
		versions = append(versions, version)
	}
	slices.Sort(versions)
	for i, version := range versions {
		if version != int64(i) {
			return fmt.Errorf("embedded MySQL migration history is incomplete at version %d", i)
		}
	}
	expected := int64(handlers.ExpectedSchemaVersion)
	if versions[len(versions)-1] != expected {
		return fmt.Errorf("embedded MySQL migrations and snapshot version %d differ", expected)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var latest int64
	var applied bool
	err = tx.QueryRowContext(ctx, "SELECT version_id, is_applied FROM goose_db_version ORDER BY id DESC LIMIT 1 FOR UPDATE").Scan(&latest, &applied)
	if err != nil {
		return fmt.Errorf("read imported version marker: %w", err)
	}
	if latest != expected || !applied {
		return fmt.Errorf("imported snapshot does not record applied version %d", expected)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM goose_db_version"); err != nil {
		return fmt.Errorf("replace imported version markers: %w", err)
	}
	for _, version := range versions {
		if _, err := tx.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)", version); err != nil {
			return fmt.Errorf("record snapshot version %d: %w", version, err)
		}
	}
	return tx.Commit()
}
