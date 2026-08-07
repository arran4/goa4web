# Goose Integration Report

## Is this an actual improvement?
Yes, integrating `github.com/pressly/goose` brings several immediate and long-term improvements over the previous custom-built migration engine:
1. **Delegation of Complexity:** Goose natively manages parsing, sorting, and executing migrations within transactions. The custom solution required custom parsers, custom loop logic, and custom file naming checks. We were able to delete `internal/app/dbstart/filename.go` and `internal/app/dbstart/migrate_test.go` and significantly simplify `migrate.go` and `automigrate.go`.
2. **Standardization:** Developers are typically already familiar with Goose annotations (e.g., `-- +goose Up` and optionally `-- +goose Down` if rollbacks are desired). This lowers the learning curve for new contributors compared to learning a bespoke system. Note that we do not use `-- +goose Down` since rollbacks are not required for this project structure.
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
  1. Organize migrations into separate dialect directories: e.g. `migrations/mysql/` and `migrations/sqlite3/`.
  2. In `internal/app/dbstart/migrate.go`, configure Goose to point to the correct sub-directory `fs.Sub(f, "sqlite3")` and use `goose.SetDialect("sqlite3")`.
  - *Note on multi-dialect folders in Goose:* Goose collects migrations sequentially across all files in the base directory. Dialects sharing the same version numbers (e.g. `0001_mysql.sql` and `0001_sqlite3.sql`) will cause "duplicate version collisions" and panic. Therefore, the standard practice for multiple dialects in Goose is dialect-separated folders passed cleanly into `goose.SetBaseFS(f)`.

## Summary
The migration to Goose simplifies code maintenance, standardizes the development flow, and builds a robust foundation for multi-database expansion. The initial pain of adapting legacy data schemas and file naming conventions is vastly outweighed by removing brittle, custom-built SQL parsing logic.