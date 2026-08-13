# database

## Purpose

Package `database` manages the application's core data storage strategies and query bindings for `database`. This involves sqlc-generated structs and interfaces.

## Why It Exists

To centralize the schema definitions and test data. This directory acts as the single source of truth for the physical database structure.

## What It Allows

It allows easy deployment of new environments and reliable unit testing against a known, consistent schema.

## Structure and Components

The core schema definitions (`schema.mysql.sql`) and testing seeds (`seed.sql`) are maintained here. SQL queries for use with `sqlc` should generally be placed under `internal/db/`.

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/database"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
