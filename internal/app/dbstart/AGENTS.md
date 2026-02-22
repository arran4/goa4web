# Database Schema Changes

Any change to the database schema (e.g., adding tables, columns, indexes, or modifying existing structures) **must** be accompanied by a new migration file in the `migrations/` directory.

- Migration files should be named sequentially (e.g., `0084.mysql.sql`).
- Ensure the migration is idempotent if possible, or handles potential conflicts gracefully.
- Update the `ExpectedSchemaVersion` in `handlers/constants.go` to match the new migration version.
