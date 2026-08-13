# core

## Purpose

Package `core` contains foundational business logic and shared utilities for `core` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `session.go`
- `fs.go`
- `memfs_testutil.go`

### Exported Types and Interfaces

- **`OSDirFS`**:
  - Methods: `MkdirAll`, `Stat`, `Remove`
- **`ContextValues`**:
- **`FileSystem`** (Interface): Defines a core contract for this module.
- **`DirFS`** (Interface): Defines a core contract for this module.
- **`OSFS`**:
  - Methods: `ReadFile`, `WriteFile`

### Exported Functions

- `GetSession`
- `GetSessionOrFail`
- `SessionErrorRedirect`
- `SessionError`
- `UseMemFS`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
