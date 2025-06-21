# Development Guidelines

Configuration values may be supplied in three ways and must be resolved in the following order of precedence:

1. Command line flags
2. Values from a config file
3. Environment variables

Defaults should only be used when a value is still empty after applying the above rules. See `email.go` for an example of this pattern.

All const declarations should include a short comment describing their purpose.

Environment variable names are centralised in `config/env.go`.

Tests must not interact with the real file system. Use in-memory file systems
provided by the `io/fs` package or mocks when file access is required.

SQL query files are compiled using `sqlc`. Do not manually edit the generated
`*.sql.go` files; instead update the corresponding `.sql` file and run `sqlc generate`.

All database schema changes must include a migration script in the `migrations/`
directory so existing installations can be upgraded.

- Errors in critical functions like main() or run() must be logged or wrapped using fmt.Errorf with context. Prefer doing both when the error propagates.

