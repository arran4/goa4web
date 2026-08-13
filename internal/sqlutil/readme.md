# internal/sqlutil

## Purpose

Package `sqlutil` provides internal, non-exported utilities and service integrations specific to `sqlutil`.

## Structure and Components

The primary files and their general responsibilities include:

- `run_statements.go`

### Exported Functions

- `RunStatements`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/sqlutil"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
