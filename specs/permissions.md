# Permissions Overview

This document describes how Goa4Web evaluates permissions using roles and grants.

## Roles

Roles group high-level capabilities. Standard roles: `anyone`, `user`, `content writer`, `moderator`, `labeler`, `administrator`.

Key flags:
- `is_admin`: Bypasses permission checks.
- `can_login`: Allows authentication.

## Grants

Permissions are grant-based, stored in the `grants` table. A grant matches a request if:
1. `user_id` matches the requester, OR
2. `role_id` matches one of the requester's roles, OR
3. Both are `NULL` (public).

### Resolution Logic

1. **DB Check**: Queries search `grants` for a matching `section`, `item`, and `action`.
2. **Admin Override**: Users with `is_admin` roles bypass these checks (handled via `Allowed` or `HasAdminRole`).

### Common Actions

| Verb | Meaning |
|---|---|
| `see` | List/Discover |
| `view` | Read full details |
| `post` | Create new item |
| `reply` | Comment/Respond |
| `edit` | Update own item |
| `edit-any` | Update others' items |
| `delete-own` | Delete own item |
| `delete-any` | Delete others' items |
| `search` | Search within section |
| `create` | Create topic (Private Forum) |

*Note: `admin` actions exist for specific administrative overrides.*

## Configuration

Valid `section` and `item` pairs are defined in `handlers/admin/role_grants.go` (`GrantActionMap`). The UI uses this to enforce valid grants.

## Query Filtering

To ensure data security, lists are filtered in SQL using a `lister_id` (the current user). The query joins the `grants` table to ensure the user has the required action (e.g., `see` or `view`) for the specific row or section.

**Example Pattern:**
```sql
...
WHERE EXISTS (
  SELECT 1 FROM grants g
  WHERE (g.user_id = @lister_id OR g.role_id IN (...effective roles...))
    AND g.section = 'forum' AND g.action = 'see'
    ...
)
```

## Seeding & Defaults

- `seed.sql` populates initial roles.
- `anyone` role grants public read access where appropriate.
- Writers get item-specific `edit` grants upon creating content.
