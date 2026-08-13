# core/common/testdata

## Purpose

Package `testdata` contains foundational business logic and shared utilities for `testdata` that are used application-wide.

## Why It Exists

To house logic, constants, and utilities that are required universally across handlers, workers, and internal services.

## What It Allows

It prevents code duplication. For example, `CoreData` is defined here and passed everywhere to provide unified access to the database and configuration state.

## Structure and Components

The primary files and their general responsibilities include:

- `querier_stub.go`

### Exported Functions

- `VisibleThreadLabels`
- `SampleSubscriptions`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/core/common/testdata"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
