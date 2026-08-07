# Goose Integration Report

## Is this an actual improvement?
Yes, integrating `github.com/pressly/goose` brings several immediate and long-term improvements over the previous custom-built migration engine:
1. **Delegation of Complexity:** Goose natively manages parsing, sorting, and executing migrations within transactions. The custom solution required custom parsers, custom loop logic, and custom file naming checks. We were able to delete `internal/app/dbstart/filename.go` and `internal/app/dbstart/migrate_test.go` and significantly simplify `migrate.go` and `automigrate.go`.
2. **Standardization:** Developers are typically already familiar with Goose annotations (e.g., `-- +goose Up` and `-- +goose Down`). This lowers the learning curve for new contributors compared to learning a bespoke system.
3. **Automated State Tracking:** We no longer have to manually append `UPDATE schema_version SET version = X;` at the end of every single migration script. Goose handles tracking automatically via the `goose_db_version` table.

## What causes difficulty?
1. **Legacy Transition:** Converting existing systems that are actively using the `schema_version` table to the `goose_db_version` table requires care. To ensure live systems didn't break or attempt to re-run all 89 migrations, we had to implement a custom `transitionToGoose` function that intercepts the start process, detects the old legacy version, and manually seeds the `goose_db_version` table up to that version before Goose takes over.
2. **File Naming Changes:** Goose uses a strict delimiter (e.g., `_`) for parsing versions. The old system used `NNNN.mysql.sql`. We had to rename 89 files to `NNNN_mysql.sql`.
3. **Loss of Fine-Grained Statement Control:** The legacy parser executed raw SQL by splitting statements at semicolons (`;`). Goose attempts to do the same but encapsulates them in Goose-specific statement scopes if `-- +goose StatementBegin/End` aren't used for complex procedures. Most of our existing queries were simple, but if we adopt complex triggers or procedures later, Goose annotations will be strictly required.

## How will this go with multi-databases (e.g., bringing back SQLite)?
Goose significantly enhances and simplifies multi-database capabilities:
- **Out of the box drivers:** Goose has built-in dialects for Postgres, MySQL, SQLite3, MSSQL, and more.
- **Dialect Management:** The custom system required parsing out `driver` from filenames manually in `parseMigrationFilename`. Goose can naturally manage diverse SQL files.
- **Bringing back SQLite:** To bring back SQLite support, you would simply:
  1. Add SQLite migration files (e.g., `NNNN_sqlite3.sql`).
  2. In `internal/app/dbstart/migrate.go`, `goose.SetDialect("sqlite3")` will automatically pick up those files, apply them properly in SQLite transactions, and track their state in `goose_db_version`. Goose will seamlessly ignore the `_mysql.sql` files if it's not configured to use them (though typically, Goose runs *all* valid `.sql` files in the folder unless you separate directories by dialect or use Go-based migrations).
  - *Note on multi-dialect folders in Goose:* Goose generally attempts to run all `.sql` files in a given target directory sequentially. If you plan to support both MySQL and SQLite simultaneously, the best practice with Goose is to put MySQL migrations in `migrations/mysql/` and SQLite migrations in `migrations/sqlite3/`. The `f` (FS) passed to `goose.SetBaseFS(f)` would then just point to the dialect-specific sub-directory based on the current configuration.

## Summary
The migration to Goose simplifies code maintenance, standardizes the development flow, and builds a robust foundation for multi-database expansion. The initial pain of adapting legacy data schemas and file naming conventions is vastly outweighed by removing brittle, custom-built SQL parsing logic.