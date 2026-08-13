# internal/dbdrivers

## Purpose

Package `dbdrivers` encapsulates the database driver initialization and specific dialect requirements for `dbdrivers`.

## Why It Exists

To abstract the specific SQL connection logic away from the rest of the application. It handles connection pooling, dialect specifics, and initialization.

## What It Allows

It allows the application to easily switch databases (e.g., from MySQL to SQLite for local testing) without rewriting query logic or changing core implementation files.

## Structure and Components

The primary files and their general responsibilities include:

- `registry.go`

### Exported Types and Interfaces

- **`DBDriver`** (Interface): Defines a core contract for this module.
- **`Registry`**:
  - Methods: `RegisterDriver`, `Drivers`, `Connector`, `Driver`, `Backup`, `Restore`, `Names`

### Exported Functions

- `NewRegistry`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/dbdrivers"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
