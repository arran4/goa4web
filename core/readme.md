# core

## Purpose

Package `core` contains foundational business logic and shared utilities for `core` that are used application-wide.

## Context and Use Cases (How and Why)

**Why it exists:** To house logic, constants, and utilities that are required universally across handlers, workers, and internal services.
**What this allows:** It prevents code duplication. For example, `CoreData` is defined here and passed everywhere to provide unified access to the database and configuration.
**How to use it:** Import the `core/*` package and invoke its exported utilities. Avoid adding dependencies from `core` to higher-level packages like `handlers` to prevent import cycles.

## Structure and Components

The primary files and their general responsibilities include:

- `fs.go`
- `memfs_testutil.go`
- `session.go`

### Exported Types and Interfaces

- **`ContextValues`**:
- **`FileSystem`** (Interface): Defines a core contract for this module.
- **`DirFS`** (Interface): Defines a core contract for this module.
- **`OSFS`**:
  - Methods: `ReadFile`, `WriteFile`
- **`OSDirFS`**:
  - Methods: `MkdirAll`, `Stat`, `Remove`

### Exported Functions

- `UseMemFS`
- `GetSession`
- `GetSessionOrFail`
- `SessionErrorRedirect`
- `SessionError`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
