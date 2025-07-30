# Permissions Overview

This document describes how Goa4Web evaluates permissions using roles and grants.

## Roles

Roles define high level capabilities that can be assigned to users. The standard roles are:

- **anonymous** – guests who are not signed in
- **user** – regular authenticated user
- **content writer** – may publish blogs and writings
- **moderator** – moderation abilities
- **administrator** – full access

Each role includes the following flags:

- **can_login** – whether accounts assigned the role are permitted to authenticate
- **is_admin** – marks administrator roles that bypass permission checks
- **public_profile_allowed_at** – when set, users with this role may expose a public profile

Users can hold multiple roles through the `user_roles` table. Role inheritance is
modelled via entries in the `grants` table with `section = 'role'`. For example,
`administrator` inherits `moderator` and `content writer` which in turn inherit
`user`.

The `user` role does not need to be explicitly assigned. Any authenticated
account automatically gains the `user` role while the `anonymous` role applies
to every connection regardless of login state.

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
| `section`  | Permission area such as `forum`, `news`, `writing`, `imagebbs` or `role`   |
| `item`     | Optional item type (e.g. `topic`, `article`)                    |
| `rule_type`| Type of rule, typically `allow` or `deny`                      |
| `item_id`  | Optional item identifier                                       |
| `item_rule`| String rule for ranges or patterns                             |
| `action`   | Operation name such as `see` or `edit`                         |
| `extra`    | Additional parameters                                          |
| `active`   | Whether the rule is in effect                                  |

### Common Actions

Permission actions describe groups of related operations. The main verbs are:

- **see** – list or otherwise discover the item
- **view** – display the item’s full details
- **comment** – add a comment
- **reply** – respond in an existing thread
- **post** – create a new thread, blog post or article
- **edit** – update an item; writers receive an item-specific grant so they can update their own posts
- **edit-any** – update items created by others
- **delete-own** – remove a resource the user created
- **delete-any** – remove resources created by others
- **admin** – perform administrative tasks on the item
- **lock** – close a thread for further replies
- **pin** – highlight or sticky a thread
- **move** – relocate an item to another area
- **invite** – invite additional users
- **approve** – mark content as reviewed or published
- **moderate** – hide or otherwise flag inappropriate content
- **search** – run a search query

Sections may introduce extra actions but these form the base vocabulary used by
the templates and permission checks.

Each section may define additional actions, but these are the core verbs used by the system.

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

Some features store per-object defaults using role identifiers. Per-writing permissions are now represented using the `grants` table and no longer rely on a dedicated `writing_user_permissions` table. `imagebbs` boards are secured via grants in the `imagebbs` section using the `board` item type.

Legacy tables `topic_permissions` and `user_topic_permissions` were replaced by
equivalent rows in the `grants` table. Forum access is now fully controlled via
grants.

These tables reference roles via `role_id` columns instead of legacy `level`
fields.

## Seeding

The `seed.sql` file provides initial roles and establishes the role hierarchy.
It also grants the `user` role to existing accounts lacking a role.

## CLI Tools

Comment-related CLI commands accept a `--user` flag to evaluate permissions for
a specific viewer. The underlying queries now take a single `viewer_id` rather
than separate user ID parameters.

### Default News Grants

The migrations seed baseline rules for the `news` section:

| Role | Action | Item | Description |
|------|-------|------|-------------|
| `anonymous` | `see`, `view` | `post` | browse published news |
| `user` | `comment`, `reply` | `post` | participate in discussions |
| `content writer`, `administrator` | `post` | `post` | create new entries |
| `content writer` | `edit` | `post` | update own news post via item-specific grant |
| `administrator` | `edit` | `post` | update any news post |

When a writer publishes a post they automatically receive an `edit` grant tied to that post, effectively granting them update rights for that item.

Other content sections such as blogs and writings follow the same pattern: authors can post entries and receive item-scoped `edit` grants while administrators hold broader `edit` privileges.

### SQL Query Filtering

Many queries now filter results directly in SQL using `viewer_id` together with the viewer's effective roles. Each query matches against grants so only records the viewer may access are returned. The table below lists the combinations used for each section.

| `section` | `item` | `action` | `grants.user_id` | `grants.role_id` | Applies when |
|-----------|-------|---------|------------------|------------------|--------------|
| `news` | `post` | `see` or `view` | `viewer_id` | `NULL` | grant specific to that user |
| `news` | `post` | `see` or `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `news` | `post` | `see` or `view` | `NULL` | viewer role ID | role-based grant |
| `news` | `post` | `see` or `view` | `NULL` | `NULL` | public grant for everyone |
| `news` | `post` | `post` | `viewer_id` | `NULL` | grant specific to that user |
| `news` | `post` | `post` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `news` | `post` | `post` | `NULL` | viewer role ID | role-based grant |
| `news` | `post` | `post` | `NULL` | `NULL` | public grant for everyone |
| `news` | `post` | `edit` | `viewer_id` | `NULL` | grant specific to that user |
| `news` | `post` | `edit` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `news` | `post` | `edit` | `NULL` | viewer role ID | role-based grant |
| `news` | `post` | `edit` | `NULL` | `NULL` | public grant for everyone |
| `blogs` | `entry` | `see` | `viewer_id` | `NULL` | grant specific to that user |
| `blogs` | `entry` | `see` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `blogs` | `entry` | `see` | `NULL` | viewer role ID | role-based grant |
| `blogs` | `entry` | `see` | `NULL` | `NULL` | public grant for everyone |
| `blogs` | `entry` | `post` | `viewer_id` | `NULL` | grant specific to that user |
| `blogs` | `entry` | `post` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `blogs` | `entry` | `post` | `NULL` | viewer role ID | role-based grant |
| `blogs` | `entry` | `post` | `NULL` | `NULL` | public grant for everyone |
| `writing` | `article` | `see` or `view` | `viewer_id` | `NULL` | grant specific to that user |
| `writing` | `article` | `see` or `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `writing` | `article` | `see` or `view` | `NULL` | viewer role ID | role-based grant |
| `writing` | `article` | `see` or `view` | `NULL` | `NULL` | public grant for everyone |
| `writing` | `article` | `post` | `viewer_id` | `NULL` | grant specific to that user |
| `writing` | `article` | `post` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `writing` | `article` | `post` | `NULL` | viewer role ID | role-based grant |
| `writing` | `article` | `post` | `NULL` | `NULL` | public grant for everyone |
| `writing` | `article` | `edit` | `viewer_id` | `NULL` | grant specific to that user |
| `writing` | `article` | `edit` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `writing` | `article` | `edit` | `NULL` | viewer role ID | role-based grant |
| `writing` | `article` | `edit` | `NULL` | `NULL` | public grant for everyone |
| `writing` | `category` | `see` | `viewer_id` | `NULL` | grant specific to that user |
| `writing` | `category` | `see` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `writing` | `category` | `see` | `NULL` | viewer role ID | role-based grant |
| `writing` | `category` | `see` | `NULL` | `NULL` | public grant for everyone |
| `writing` | `category` | `view` | `viewer_id` | `NULL` | grant specific to that user |
| `writing` | `category` | `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `writing` | `category` | `view` | `NULL` | viewer role ID | role-based grant |
| `writing` | `category` | `view` | `NULL` | `NULL` | public grant for everyone |
| `imagebbs` | `board` | `see` or `view` | `viewer_id` | `NULL` | grant specific to that user |
| `imagebbs` | `board` | `see` or `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `imagebbs` | `board` | `see` or `view` | `NULL` | viewer role ID | role-based grant |
| `imagebbs` | `board` | `see` or `view` | `NULL` | `NULL` | public grant for everyone |
| `imagebbs` | `board` | `post` | `viewer_id` | `NULL` | grant specific to that user |
| `imagebbs` | `board` | `post` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `imagebbs` | `board` | `post` | `NULL` | viewer role ID | role-based grant |
| `imagebbs` | `board` | `post` | `NULL` | `NULL` | public grant for everyone |
| `linker` | `category` | `see` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `category` | `see` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `category` | `see` | `NULL` | viewer role ID | role-based grant |
| `linker` | `category` | `see` | `NULL` | `NULL` | public grant for everyone |
| `linker` | `category` | `view` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `category` | `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `category` | `view` | `NULL` | viewer role ID | role-based grant |
| `linker` | `category` | `view` | `NULL` | `NULL` | public grant for everyone |
| `forum` | `category` | `see` | `viewer_id` | `NULL` | grant specific to that user |
| `forum` | `category` | `see` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `forum` | `category` | `see` | `NULL` | viewer role ID | role-based grant |
| `forum` | `category` | `see` | `NULL` | `NULL` | public grant for everyone |
| `forum` | `category` | `view` | `viewer_id` | `NULL` | grant specific to that user |
| `forum` | `category` | `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `forum` | `category` | `view` | `NULL` | viewer role ID | role-based grant |
| `forum` | `category` | `view` | `NULL` | `NULL` | public grant for everyone |
| `linker` | `link` | `see` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `link` | `see` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `link` | `see` | `NULL` | viewer role ID | role-based grant |
| `linker` | `link` | `see` | `NULL` | `NULL` | public grant for everyone |
| `linker` | `link` | `view` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `link` | `view` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `link` | `view` | `NULL` | viewer role ID | role-based grant |
| `linker` | `link` | `view` | `NULL` | `NULL` | public grant for everyone |
| `linker` | `link` | `comment` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `link` | `comment` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `link` | `comment` | `NULL` | viewer role ID | role-based grant |
| `linker` | `link` | `reply` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `link` | `reply` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `link` | `reply` | `NULL` | viewer role ID | role-based grant |
| `linker` | `link` | `post` | `viewer_id` | `NULL` | grant specific to that user |
| `linker` | `link` | `post` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `linker` | `link` | `post` | `NULL` | viewer role ID | role-based grant |
| `forum` | `topic` | `see` | `viewer_id` | `NULL` | grant specific to that user |
| `forum` | `topic` | `see` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `forum` | `topic` | `see` | `NULL` | viewer role ID | role-based grant |
| `forum` | `topic` | `see` | `NULL` | `NULL` | public grant for everyone |
| `forum` | `topic` | `post` | `viewer_id` | `NULL` | grant specific to that user |
| `forum` | `topic` | `post` | `viewer_id` | viewer role ID | grant requiring both user and role |
| `forum` | `topic` | `post` | `NULL` | viewer role ID | role-based grant |
| `forum` | `topic` | `post` | `NULL` | `NULL` | public grant for everyone |
| `admin` | `page` | `view` | `viewer_id` | `NULL` | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `view` | `viewer_id` | viewer role ID | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `view` | `NULL` | viewer role ID | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `view` | `NULL` | `NULL` | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `edit` | `viewer_id` | `NULL` | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `edit` | `viewer_id` | viewer role ID | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `edit` | `NULL` | viewer role ID | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `edit` | `NULL` | `NULL` | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `admin` | `viewer_id` | `NULL` | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `admin` | `viewer_id` | viewer role ID | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `admin` | `NULL` | viewer role ID | access requires `administrator` role via `AdminCheckerMiddleware` |
| `admin` | `page` | `admin` | `NULL` | `NULL` | access requires `administrator` role via `AdminCheckerMiddleware` |

Administrator endpoints are guarded by the `AdminCheckerMiddleware` implemented
in `internal/router/router.go`. The middleware calls `corecommon.Allowed`, which
loads roles for the current user using the `GetPermissionsByUserID` query from
`internal/db/queries-permissions.sql`. Only users with the `administrator` role
can reach these routes.

Listing pages and RSS feeds still invoke `HasGrant` on each row for extra safety.

The same lookup pattern covers every action. Whether a user wants to `post`,
`reply`, `comment`, `write` or `edit`, the query matches grants by `viewer_id`
and roles using the unified action names above. This ensures consistent
terminology across sections and avoids duplicated rules.
