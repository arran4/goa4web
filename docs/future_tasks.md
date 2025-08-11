# Future Improvements

- Rewrite older PostgreSQL and SQLite migration scripts to use engine-appropriate syntax instead of MySQL-specific `CHANGE COLUMN` directives.
- Regenerate `schema/*.sql` files for PostgreSQL and SQLite to reflect dialect-specific features.
- Add automated tests exercising migration application on PostgreSQL and SQLite backends.
- Create lint checks to prevent introduction of MySQL-only syntax into non-MySQL migration files.
- Improve template coverage with tests for additional rendering contexts and data structures.
