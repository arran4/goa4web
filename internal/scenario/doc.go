// Package scenario implements a human-readable scenario/seeding system for goa4web.
//
// # Core Philosophy
//
// "The scenario describes what happened; current goa4web decides how that intent is stored."
//
// Scenarios are human-readable, durable journals of application events (user creations,
// role assignments, forum discussions, blog posts, etc.) formatted in TXTAR.
//
// Scenarios represent application/domain intent rather than physical SQL table fixtures,
// dumps, or schema-version-bound data. When an older scenario is replayed against a newer
// version of goa4web, it uses the current schema, UUID/ID generation, password hashing,
// image pipelines, and business semantics.
//
// Scenarios are intended for creating valid manual and automated test environments and seeds.
// Historical database migration testing remains a separate concern using old database snapshots
// and migration test fixtures.
//
// # Format Structure
//
// A scenario is defined in a TXTAR archive containing:
//   - `scenario.meta`: Format identifier (e.g. `Format: goa4web-scenario/v1`) and scenario name.
//   - Ordered `*.event` members: Each event defines an application operation (`Op: <op>`),
//     a timestamp (`At: <RFC3339>`), optional declared references (`Ref: <symbol>`),
//     operation-specific headers, and an optional multiline body.
//
// Binary assets such as images are kept as adjacent files (e.g. `assets/image.jpg`)
// rather than embedded inside the TXTAR archive.
package scenario
