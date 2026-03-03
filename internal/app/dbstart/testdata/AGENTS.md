# Historical Test Data

This directory contains `original.*.sql` files which represent the initial state of the database schema (Version 1) for different dialects.

**Modification Policy:**
These files are for records and migration testing ONLY. **Do not modify these files under any circumstances.** Even if you find code health issues (like missing primary keys) or TODO comments within them, you must leave them untouched. They represent a historical snapshot of the database that our migration scripts must be able to process exactly as it was.
