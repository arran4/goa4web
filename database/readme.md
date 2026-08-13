# database

## Purpose

Package `database` manages the application's core data storage strategies and query bindings for `database`. This involves sqlc-generated structs and interfaces.

## Context and Use Cases (How and Why)

**Why it exists:** To centralize the schema definitions and test data. This is the source of truth for the database structure.
**What this allows:** It allows easy deployment of new environments and reliable unit testing against a known, consistent schema.
**How to use it:** When adding a new table, add the migration to `migrations/` and manually reflect the final state in `database/schema.mysql.sql`.

## Structure and Components

The core schema definitions (`schema.mysql.sql`) and testing seeds (`seed.sql`) are maintained here. SQL queries for use with `sqlc` should generally be placed under `internal/db/`.

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/database"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
