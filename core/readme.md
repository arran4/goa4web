# core

## Purpose

Package `core` contains foundational business logic and shared utilities for `core` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `memfs_testutil.go`
- `session.go`
- `fs.go`

### Exported Types

- `ContextValues`
- `FileSystem`
- `DirFS`
- `OSFS`
- `OSDirFS`

### Exported Functions

- `UseMemFS`
- `GetSession`
- `GetSessionOrFail`
- `SessionErrorRedirect`
- `SessionError`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
