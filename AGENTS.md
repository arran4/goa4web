# Development Guidelines

This repository powers the Goa4Web services. Follow these conventions when modifying the code base.

## Specifications

The `specs/` directory documents the current implementation and architecture. These files are the source of truth for understanding the system.

- **Reflection**: The specs reflect the current code state.
- **Updates**: Changes to specification files should only be made when explicitly requested via a prompt.

Refer to `specs/query_naming.md` for SQL naming conventions and `specs/permissions.md` for the permissions model.

## Database and migrations

These rules intentionally live at repository scope because a database change often
touches `internal/db`, `database`, `migrations`, handlers, and application code:

- Add application queries to `internal/db/*.sql`, run `sqlc generate`, and never
  edit generated `*.sql.go` files by hand. Prefer the generated `db.Querier`
  interface over exposing `*sql.DB`, and proxy queries through `CoreData` rather
  than calling `cd.Queries()` so caching and invalidation remain possible.
- Every schema change requires a new numbered migration. Never edit an existing
  migration: deployed installations may already have applied it.
- Keep `database/schema.mysql.sql` synchronized with every schema migration, and
  update the migration's `schema_version` plus `ExpectedSchemaVersion` in
  `handlers/constants.go`.
- Keep roles introduced by migrations synchronized with `database/seed.sql`.

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


Errors in critical functions like `main()` or `run()` must be logged or wrapped using `fmt.Errorf` with context. Prefer doing both when errors propagate.

All default HTML or text templates must exist as standalone files and be embedded using `//go:embed` rather than inline string constants.

Forum page templates that are parsed by filename (e.g., `core/templates/site/forum/adminTopicsPage.gohtml`) should not wrap the entire file in a redundant `{{ define "forum/<filename>" }}` block.

When tackling bugs or missing features, check if the behaviour can be verified with tests. If so, write a test that fails before changing the implementation. Iterate on your fix until the new test passes.

Before committing, run `go mod tidy` followed by `go fmt ./...`, `go vet ./...`, and `golangci-lint` to match the CI checks. If `go mod tidy` fails, continue but mention the error in the PR summary.

Do not add new global variables unless explicitly instructed or already well established. Avoid global state. Use dependency injection (e.g., passing structs via options/constructors) to share state like caches.

## Database and Testing Notes

- If the database setup is blocking frontend verification, it is acceptable to skip it and note that the user may perform manual testing instead.

## Quality Assurance

After every successful build or significant code changeset, you must run the following commands to ensure code quality:

```bash
go fmt ./...
go vet ./...
go test ./...
```

## Verification Tooling

A CLI tool is available to verify template rendering with mock data. This is useful for generating HTML snapshots for testing or visual verification without running the full server.

When working on HTML changes, use this tool with the `-listen` flag to start a local server and verify the visual output. You are encouraged to capture screenshots of the changes and share them in the chat to confirm the results.

**Note:** Do not commit these verification screenshots to the repository unless they are useful for the README or a documentation gallery. Please also do not commit generated binary files such as `server_bin`.

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

## Directory Structure

- **a4code**
  - **a4code/a4code2html**
  - **a4code/ast**
  - **a4code/format**
    - **a4code/format/test**
  - **a4code/goa4webhtml**
    - **a4code/goa4webhtml/test**
  - **a4code/html**
    - **a4code/html/test**
  - **a4code/markdown**
    - **a4code/markdown/test**
  - **a4code/text**
    - **a4code/text/test**
- **cmd**
  - **cmd/a4code2html**
  - **cmd/gen-permutations**
  - **cmd/goa4web**
- **config**
- **core**
  - **core/common**
    - **core/common/testdata**
  - **core/consts**
  - **core/language**
  - **core/templates**
- **database**
- **handlers**
  - **handlers/admin**
  - **handlers/admincommon**
  - **handlers/auth**
  - **handlers/blogs**
  - **handlers/bookmarks**
  - **handlers/externallink**
  - **handlers/faq**
  - **handlers/forum**
    - **handlers/forum/comments**
  - **handlers/handlertest**
  - **handlers/imagebbs**
  - **handlers/images**
  - **handlers/languages**
  - **handlers/linker**
  - **handlers/news**
  - **handlers/privateforum**
  - **handlers/search**
  - **handlers/share**
  - **handlers/user**
  - **handlers/writings**
- **internal**
  - **internal/adminapi**
  - **internal/algorithms**
  - **internal/app**
    - **internal/app/dbstart**
    - **internal/app/server**
  - **internal/configexplain**
  - **internal/configformat**
  - **internal/db**
  - **internal/dbdrivers**
    - **internal/dbdrivers/dbdefaults**
    - **internal/dbdrivers/mysql**
  - **internal/dbops**
  - **internal/dlq**
    - **internal/dlq/db**
    - **internal/dlq/dir**
    - **internal/dlq/dlqdefaults**
    - **internal/dlq/email**
    - **internal/dlq/file**
    - **internal/dlq/mock**
  - **internal/email**
    - **internal/email/emaildefaults**
    - **internal/email/jmap**
    - **internal/email/local**
    - **internal/email/log**
    - **internal/email/mock**
    - **internal/email/sendgrid**
    - **internal/email/ses**
    - **internal/email/smtp**
  - **internal/eventbus**
  - **internal/faq_templates**
  - **internal/images**
  - **internal/middleware**
    - **internal/middleware/apiauth**
    - **internal/middleware/csrf**
  - **internal/navigation**
  - **internal/notifications**
  - **internal/opengraph**
  - **internal/permissions**
  - **internal/role_templates**
  - **internal/roles**
  - **internal/router**
  - **internal/scheduler**
  - **internal/secrets**
  - **internal/sign**
    - **internal/sign/signutil**
  - **internal/sqlutil**
  - **internal/stats**
  - **internal/subscription_templates**
  - **internal/subscriptions**
  - **internal/tasks**
  - **internal/testhelpers**
  - **internal/upload**
    - **internal/upload/local**
    - **internal/upload/s3**
    - **internal/upload/uploaddefaults**
  - **internal/websocket**
- **migrations**
- **workers**
  - **workers/auditworker**
  - **workers/backgroundtaskworker**
  - **workers/emailqueue**
  - **workers/externallinkworker**
  - **workers/logworker**
  - **workers/postcountworker**
  - **workers/searchworker**
