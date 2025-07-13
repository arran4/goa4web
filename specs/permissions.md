# Permissions Overview

This document describes how Goa4Web evaluates permissions using roles and grants.

## Roles

Roles define high level capabilities that can be assigned to users. The standard roles are:

- **anonymous** – guests who are not signed in
- **user** – regular authenticated user
- **content writer** – may publish blogs and writings
- **moderator** – moderation abilities
- **administrator** – full access

Users can hold multiple roles through the `user_roles` table. Role inheritance is
modelled via entries in the `grants` table with `section = 'role'`. For example,
`administrator` inherits `moderator` and `content writer` which in turn inherit
`user`.

## Grants Table

All permission rules live in the `grants` table. A grant may apply to a specific
user (`user_id`), a role (`role_id`) or everyone (both columns `NULL`). The
columns are:

| Column     | Purpose                                                        |
|------------|----------------------------------------------------------------|
| `id`       | Primary key                                                    |
| `created_at`, `updated_at` | Timestamps                                     |
| `user_id`  | Optional user the rule targets                                 |
| `role_id`  | Optional role the rule targets                                 |
| `section`  | Permission area such as `forum`, `news`, `writing` or `role`   |
| `item`     | Optional item type (e.g. `topic`, `article`)                    |
| `rule_type`| Type of rule, typically `allow` or `deny`                      |
| `item_id`  | Optional item identifier                                       |
| `item_rule`| String rule for ranges or patterns                             |
| `action`   | Operation like `see`, `post`, `edit`                           |
| `extra`    | Additional parameters                                          |
| `active`   | Whether the rule is in effect                                  |

## Permission Resolution

When checking access for a user, the application resolves the user's effective
roles using the `ListEffectiveRoleIDsByUserID` query. This query recursively
expands role inheritance defined in `grants`.

A permission check searches the `grants` table for an active row matching the
requested `section`, `item` and `action`. A grant is considered applicable if
one of the following is true:

1. `user_id` matches the user making the request
2. `role_id` is one of the user's effective role IDs
3. both `user_id` and `role_id` are `NULL` (applies to everyone)

If no matching grant is found, the action is denied unless the user has the
`administrator` role which bypasses checks.

## Additional Tables

Some features store per-object defaults using role identifiers:

- `topic_permissions` – default required roles for forum topics
- `user_topic_permissions` – user specific topic rules
- `writing_user_permissions` – per writing user access

These tables reference roles via `role_id` columns instead of legacy `level`
fields.

## Seeding

The `seed.sql` file provides initial roles and establishes the role hierarchy.
It also grants the `user` role to existing accounts lacking a role.

