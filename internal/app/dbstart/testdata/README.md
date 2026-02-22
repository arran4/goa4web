# Test Data

This directory contains `original.*.sql` files which represent the initial state of the database schema (Version 1) for different dialects.

**Purpose:**
These files are used by migration tests to verify that the migration system can correctly upgrade a database starting from this initial state to the current schema version.

**Modification Policy:**
- Do not modify these files to reflect current schema changes; they are meant to be historical snapshots.
- Exceptions can be made for correcting clear errors or code health issues (e.g., missing primary keys in the original definition) that do not invalidate the migration path.
