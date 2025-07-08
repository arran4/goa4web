# Development Guidelines

This repository powers the Goa4Web services. Follow these conventions when modifying the code base.

Configuration values may be supplied in three ways and must be resolved in this order:

1. Command line flags
2. Values from a config file
3. Environment variables

Defaults should only be used when a value is still empty after applying the above rules. See `runtimeconfig.GenerateRuntimeConfig` for details.

Environment variable names are centralised in `config/env.go`. Example configuration files live in `examples/` and use the same keys.

All `const` declarations must include a short comment describing their purpose.

Tests must not interact with the real file system. Use in-memory file systems from `io/fs` or mocks when file access is required. Run all tests with:

```
go test ./...
```

SQL query files are compiled using `sqlc`. Do not manually edit the generated `*.sql.go` files; instead edit the `.sql` files under `internal/db/` and run `sqlc generate`.

All database schema changes must include a new migration script in the `migrations/` directory (for example `0002.sql`, `0003.sql`). Never modify existing migration files as that would break deployments running older versions.

Errors in critical functions like `main()` or `run()` must be logged or wrapped using `fmt.Errorf` with context. Prefer doing both when errors propagate.

All default HTML or text templates must exist as standalone files and be embedded using `//go:embed` rather than inline string constants.

When tackling bugs or missing features, check if the behaviour can be verified with tests. If so, write a test that fails before changing the implementation. Iterate on your fix until the new test passes.

Before committing, run `go mod tidy` followed by `go fmt ./...`, `go vet ./...`, and `golangci-lint` to match the CI checks. If `go mod tidy` fails, continue but mention the error in the PR summary.
