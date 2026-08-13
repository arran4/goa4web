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

- **a4code**: Package `a4code` is the root package for the custom A4Code markup engine. It defines the core parser, tokenization, and entry points for evaluating A4Code strings.
  - **a4code/a4code2html**: Package a4code2html converts a small markup language into HTML or
  - **a4code/ast**: Package `ast` defines the Abstract Syntax Tree (AST) nodes for the A4Code markup language. It provides the core data structures used to represent parsed A4Code elements in memory before they are formatted or rendered.
  - **a4code/format**: Package `format` provides utilities for taking an A4Code Abstract Syntax Tree (AST) and formatting it back into a valid, normalized A4Code string. This is useful for pretty-printing or normalizing user input.
    - **a4code/format/test**: Package `format` provides utilities for taking an A4Code Abstract Syntax Tree (AST) and formatting it back into a valid, normalized A4Code string. This is useful for pretty-printing or normalizing user input.
  - **a4code/goa4webhtml**: Package `goa4webhtml` provides specialized HTML rendering for A4Code that is specifically tailored and integrated with the Goa4Web templating and asset system.
    - **a4code/goa4webhtml/test**: Package `goa4webhtml` provides specialized HTML rendering for A4Code that is specifically tailored and integrated with the Goa4Web templating and asset system.
  - **a4code/html**: Package `html` provides the rendering engine that converts an A4Code Abstract Syntax Tree (AST) into standard HTML output suitable for web browsers.
    - **a4code/html/test**: Package `html` provides the rendering engine that converts an A4Code Abstract Syntax Tree (AST) into standard HTML output suitable for web browsers.
  - **a4code/markdown**: Package `markdown` provides utilities for converting standard Markdown input into A4Code markup, or potentially rendering A4Code as Markdown.
    - **a4code/markdown/test**: Package `markdown` provides utilities for converting standard Markdown input into A4Code markup, or potentially rendering A4Code as Markdown.
  - **a4code/text**: Package `text` provides a plain-text renderer for the A4Code Abstract Syntax Tree (AST), useful for stripping formatting and extracting pure content.
    - **a4code/text/test**: Package `text` provides a plain-text renderer for the A4Code Abstract Syntax Tree (AST), useful for stripping formatting and extracting pure content.
  - **cmd/a4code2html**: Package `main` defines a main executable entry point for the `a4code2html` application or CLI tool.
  - **cmd/gen-permutations**: Package `main` defines a main executable entry point for the `gen-permutations` application or CLI tool.
  - **cmd/goa4web**: Package `main` defines a main executable entry point for the `goa4web` application or CLI tool.
- **config**: Package `config` defines the data structures and parsing logic for the Goa4Web application configuration. It handles reading from environment variables, command-line flags, and configuration files.
- **core**: Package `core` contains foundational business logic and shared utilities for `core` that are used application-wide.
  - **core/common**: Package `common` contains foundational business logic and shared utilities for `common` that are used application-wide.
    - **core/common/testdata**: Package `testdata` contains foundational business logic and shared utilities for `testdata` that are used application-wide.
  - **core/consts**: Package `consts` contains foundational business logic and shared utilities for `consts` that are used application-wide.
  - **core/language**: Package `language` contains foundational business logic and shared utilities for `language` that are used application-wide.
  - **core/templates**: Package `templates` contains foundational business logic and shared utilities for `templates` that are used application-wide.
- **database**: Package `database` manages the application's core data storage strategies and query bindings for `database`. This involves sqlc-generated structs and interfaces.
- **handlers**: The `handlers` package and its subdirectories encompass the web presentation layer for Goa4Web. This is where HTTP requests are received, authorized, routed to specific logical sub-handlers, and responded to. It is the primary entry point for user interaction via the web interface. Things that should become handlers: new API routes, page views, and form submission endpoints.
  - **handlers/admin**: Package `admin` handles HTTP requests for the `admin` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/admincommon**: Package `admincommon` handles HTTP requests for the `admincommon` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/auth**: Package `auth` handles HTTP requests for the `auth` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/blogs**: Package `blogs` handles HTTP requests for the `blogs` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/bookmarks**: Package `bookmarks` handles HTTP requests for the `bookmarks` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/externallink**: Package `externallink` handles HTTP requests for the `externallink` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/faq**: Package `faq` handles HTTP requests for the `faq` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/forum**: Package `forum` handles HTTP requests for the `forum` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
    - **handlers/forum/comments**: Package `comments` handles HTTP requests for the `comments` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/handlertest**: Package `handlertest` handles HTTP requests for the `handlertest` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/imagebbs**: Package `imagebbs` handles HTTP requests for the `imagebbs` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/images**: Package `images` handles HTTP requests for the `images` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/languages**: Package `languages` handles HTTP requests for the `languages` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/linker**: Package `linker` handles HTTP requests for the `linker` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/news**: Package `news` handles HTTP requests for the `news` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/privateforum**: Package `privateforum` handles HTTP requests for the `privateforum` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/search**: Package `search` handles HTTP requests for the `search` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/share**: Package `share` handles HTTP requests for the `share` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/user**: Package `user` handles HTTP requests for the `user` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **handlers/writings**: Package `writings` handles HTTP requests for the `writings` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.
  - **internal/adminapi**: Package `adminapi` provides internal, non-exported utilities and service integrations specific to `adminapi`.
  - **internal/algorithms**: Package `algorithms` provides internal, non-exported utilities and service integrations specific to `algorithms`.
  - **internal/app**: Package `app` provides internal, non-exported utilities and service integrations specific to `app`.
    - **internal/app/dbstart**: Package `dbstart` provides internal, non-exported utilities and service integrations specific to `dbstart`.
    - **internal/app/server**: Package `server` provides internal, non-exported utilities and service integrations specific to `server`.
  - **internal/configexplain**: Package `configexplain` provides internal, non-exported utilities and service integrations specific to `configexplain`.
  - **internal/configformat**: Package `configformat` provides internal, non-exported utilities and service integrations specific to `configformat`.
  - **internal/db**: Package `db` provides internal, non-exported utilities and service integrations specific to `db`.
  - **internal/dbdrivers**: Package `dbdrivers` encapsulates the database driver initialization and specific dialect requirements for `dbdrivers`.
    - **internal/dbdrivers/dbdefaults**: Package `dbdefaults` encapsulates the database driver initialization and specific dialect requirements for `dbdefaults`.
    - **internal/dbdrivers/mysql**: Package `mysql` encapsulates the database driver initialization and specific dialect requirements for `mysql`.
  - **internal/dbops**: Package `dbops` provides internal, non-exported utilities and service integrations specific to `dbops`.
  - **internal/dlq**: Package `dlq` provides internal, non-exported utilities and service integrations specific to `dlq`.
    - **internal/dlq/db**: Package `db` provides internal, non-exported utilities and service integrations specific to `db`.
    - **internal/dlq/dir**: Package `dir` provides internal, non-exported utilities and service integrations specific to `dir`.
    - **internal/dlq/dlqdefaults**: Package `dlqdefaults` provides internal, non-exported utilities and service integrations specific to `dlqdefaults`.
    - **internal/dlq/email**: Package `email` provides internal, non-exported utilities and service integrations specific to `email`.
    - **internal/dlq/file**: Package `file` provides internal, non-exported utilities and service integrations specific to `file`.
    - **internal/dlq/mock**: Package `mock` provides internal, non-exported utilities and service integrations specific to `mock`.
  - **internal/email**: The `internal/email` directory encapsulates all logic related to constructing, dispatching, and managing electronic mail within the system. It abstracts the underlying providers so the core application logic remains decoupled from specific services like AWS SES.
    - **internal/email/emaildefaults**: Package `emaildefaults` provides concrete implementations or abstractions for the `emaildefaults` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/jmap**: Package `jmap` provides concrete implementations or abstractions for the `jmap` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/local**: Package `local` provides concrete implementations or abstractions for the `local` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/log**: Package `log` provides concrete implementations or abstractions for the `log` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/mock**: Package `mock` provides concrete implementations or abstractions for the `mock` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/sendgrid**: Package `sendgrid` provides concrete implementations or abstractions for the `sendgrid` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/ses**: Package `ses` provides concrete implementations or abstractions for the `ses` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
    - **internal/email/smtp**: Package `smtp` provides concrete implementations or abstractions for the `smtp` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).
  - **internal/eventbus**: Package `eventbus` provides internal, non-exported utilities and service integrations specific to `eventbus`.
  - **internal/faq_templates**: Package `faq_templates` provides internal, non-exported utilities and service integrations specific to `faq_templates`.
  - **internal/images**: Package `images` provides internal, non-exported utilities and service integrations specific to `images`.
  - **internal/middleware**: Package `middleware` provides internal, non-exported utilities and service integrations specific to `middleware`.
    - **internal/middleware/apiauth**: Package `apiauth` provides internal, non-exported utilities and service integrations specific to `apiauth`.
    - **internal/middleware/csrf**: Package `csrf` provides internal, non-exported utilities and service integrations specific to `csrf`.
  - **internal/navigation**: Package `navigation` provides internal, non-exported utilities and service integrations specific to `navigation`.
  - **internal/notifications**: Package `notifications` provides internal, non-exported utilities and service integrations specific to `notifications`.
  - **internal/opengraph**: Package `opengraph` provides internal, non-exported utilities and service integrations specific to `opengraph`.
  - **internal/permissions**: Package `permissions` provides internal, non-exported utilities and service integrations specific to `permissions`.
  - **internal/role_templates**: Package `role_templates` provides internal, non-exported utilities and service integrations specific to `role_templates`.
  - **internal/roles**: Package `roles` provides internal, non-exported utilities and service integrations specific to `roles`.
  - **internal/router**: Package `router` provides internal, non-exported utilities and service integrations specific to `router`.
  - **internal/scheduler**: Package `scheduler` provides internal, non-exported utilities and service integrations specific to `scheduler`.
  - **internal/secrets**: Package `secrets` provides internal, non-exported utilities and service integrations specific to `secrets`.
  - **internal/sign**: Package `sign` provides internal, non-exported utilities and service integrations specific to `sign`.
    - **internal/sign/signutil**: Package `signutil` provides internal, non-exported utilities and service integrations specific to `signutil`.
  - **internal/sqlutil**: Package `sqlutil` provides internal, non-exported utilities and service integrations specific to `sqlutil`.
  - **internal/stats**: Package `stats` provides internal, non-exported utilities and service integrations specific to `stats`.
  - **internal/subscription_templates**: Package `subscriptiontemplates` provides internal, non-exported utilities and service integrations specific to `subscription_templates`.
  - **internal/subscriptions**: Package `subscriptions` provides internal, non-exported utilities and service integrations specific to `subscriptions`.
  - **internal/tasks**: Package `tasks` provides internal, non-exported utilities and service integrations specific to `tasks`.
  - **internal/testhelpers**: Package `testhelpers` provides internal, non-exported utilities and service integrations specific to `testhelpers`.
  - **internal/upload**: Package `upload` provides internal, non-exported utilities and service integrations specific to `upload`.
    - **internal/upload/local**: Package `local` provides internal, non-exported utilities and service integrations specific to `local`.
    - **internal/upload/s3**: Package `s3` provides internal, non-exported utilities and service integrations specific to `s3`.
    - **internal/upload/uploaddefaults**: Package `uploaddefaults` provides internal, non-exported utilities and service integrations specific to `uploaddefaults`.
  - **internal/websocket**: Package `websocket` provides internal, non-exported utilities and service integrations specific to `websocket`.
- **migrations**: Package `migrations` provides functionality specific to `migrations`.
- **workers**: The `workers` directory contains asynchronous background processors. A 'worker' in Goa4Web is a component that listens to the central eventbus or a queue and executes tasks outside the critical path of an HTTP request.
  - **workers/auditworker**: Package `auditworker` implements a specific background worker (`auditworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
  - **workers/backgroundtaskworker**: Package `backgroundtaskworker` implements a specific background worker (`backgroundtaskworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
  - **workers/emailqueue**: Package `emailqueue` implements a specific background worker (`emailqueue`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
  - **workers/externallinkworker**: Package `externallinkworker` implements a specific background worker (`externallinkworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
  - **workers/logworker**: Package `logworker` implements a specific background worker (`logworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
  - **workers/postcountworker**: Package `postcountworker` implements a specific background worker (`postcountworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
  - **workers/searchworker**: Package `searchworker` implements a specific background worker (`searchworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.
