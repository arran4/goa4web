# Development Guidelines

This repository powers the Goa4Web services. Follow these conventions when modifying the code base.

## Specifications

The `specs/` directory documents the current implementation and architecture. These files are the source of truth for understanding the system.

- **Reflection**: The specs reflect the current code state.
- **Updates**: Changes to specification files should only be made when explicitly requested via a prompt.

Refer to `specs/query_naming.md` for SQL naming conventions and `specs/permissions.md` for the permissions model.

## Configuration

Configuration values may be supplied in three ways and must be resolved in this order:

1. Command line flags
2. Values from a config file
3. Environment variables

Defaults should only be used when a value is still empty after applying the above rules. See `runtimeconfig.GenerateRuntimeConfig` for details.

Environment variable names are centralised in `config/env.go`. Example configuration files live in `examples/` and use the same keys.

## Coding Standards

All `const` declarations must include a short comment describing their purpose.

Tests must not interact with the real file system. Use in-memory file systems from `io/fs` or mocks when file access is required. Run all tests with:

```
go test ./...
```

To streamline the human approval process, run a full test suite with `go test ./...` instead of testing individual files or modules. This allows for a single, pre-approved overall test, which is faster for human verification.

SQL query files are compiled using `sqlc`. Do not manually edit the generated `*.sql.go` files; instead edit the `.sql` files under `internal/db/` and run `sqlc generate`.
Avoid using the `overrides` section in `sqlc.yaml`; prefer Go type aliases if a different struct name is required.

All database schema changes must include a new migration script in the `migrations/` directory (for example `0002.mysql.sql`, `0003.mysql.sql`). Never modify existing migration files as that would break deployments running older versions. Every migration must also update the `schema_version` table so deployments can track the current schema state. Bump the `ExpectedSchemaVersion` constant in `handlers/constants.go` whenever a new migration is added so tests stay in sync.

Errors in critical functions like `main()` or `run()` must be logged or wrapped using `fmt.Errorf` with context. Prefer doing both when errors propagate.

All default HTML or text templates must exist as standalone files and be embedded using `//go:embed` rather than inline string constants.

Forum page templates that are parsed by filename (e.g., `core/templates/site/forum/adminTopicsPage.gohtml`) should not wrap the entire file in a redundant `{{ define "forum/<filename>" }}` block.

When tackling bugs or missing features, check if the behaviour can be verified with tests. If so, write a test that fails before changing the implementation. Iterate on your fix until the new test passes.

Before committing, run `go mod tidy` followed by `go fmt ./...`, `go vet ./...`, and `golangci-lint` to match the CI checks. If `go mod tidy` fails, continue but mention the error in the PR summary.

Do not add new global variables unless explicitly instructed or already well established.

## Database and Testing Notes

- Roles defined in migrations must also be present in `database/seed.sql`. These two sources are strongly linked. If a role is not in the seed data (because it is optional), it should not be included in migrations.
- If the database setup is blocking frontend verification, it is acceptable to skip it and note that the user may perform manual testing instead.
- For unit tests that require a database connection, it is recommended to mock the `db.Querier` interface to avoid database dependencies.

## Quality Assurance

After every successful build or significant code changeset, you must run the following commands to ensure code quality:

```bash
go fmt ./...
go vet ./...
go test ./...
```

## Verification Tooling

A CLI tool is available to verify template rendering with mock data. This is useful for generating HTML snapshots for testing or visual verification without running the full server.

Usage:
```bash
# Render to stdout
./goa4web test verification template -template <path/to/template.gohtml> -data <data.json>

# Render to file
./goa4web test verification template -template <path/to/template.gohtml> -data <data.json> -output <file.html>

# Serve locally (serves static assets too)
./goa4web test verification template -template <path/to/template.gohtml> -data <data.json> -listen :8080
```

The JSON data file should contain the data structure expected by the template (the `Dot` field) and optional configuration:
```json
{
  "Dot": { ... },     // Data passed as {{ . }}
  "User": { ... },    // Mocked current user (optional)
  "Config": { ... },  // Runtime configuration (optional)
  "URL": "..."        // Request URL (optional)
}
```
Field types in `Dot` are automatically fixed:
- Strings in RFC3339 format are converted to `time.Time`.
- Whole number `float64` values are converted to `int32` to match typical DB IDs.
