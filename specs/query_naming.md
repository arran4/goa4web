# Query Naming Specification

Goa4Web uses `sqlc` to generate type-safe Go code from SQL. This document defines the naming conventions for queries to ensure consistent permission handling and intent.

## Prefixes & Context

Query names indicate the authorization context:

- **`System...`**: Internal operations (CLI, background tasks). **Must not** take a user ID.
- **`Admin...`**: Administrative operations. **Must not** take a user ID (auth handled by middleware/caller).
- **`List...` / `Get...`**: User-facing operations.
    - Must use a `For<Role>` suffix (e.g., `ForLister`, `ForWriter`).
    - Must take a matching `<Role>ID` parameter (e.g., `@lister_id`).
    - **Must** enforce permissions (grants) directly in the SQL `WHERE` clause.

## Conventions

1. **Pagination**: All `List` queries must include `LIMIT` and `OFFSET`.
2. **Terminology**: Use `List`/`Lister` instead of legacy `See`/`Seer`.
3. **Defense in Depth**: User queries enforce grants in SQL. Go handlers often re-check logic, but the SQL check is primary for list filtering.
4. **Custom Queries**: Avoid dynamic SQL where possible. If inevitable, use the `CustomQueries` interface, but do not expose the raw DB handle.
