# Permissions Overview

This document describes how Goa4Web evaluates permissions using roles and grants.

## Roles

Roles define high level capabilities that can be assigned to users. The standard roles are:

- **anyone** – guests who are not signed in
- **user** – regular authenticated user
- **content writer** – may publish blogs and writing articles
- **moderator** – moderation abilities
- **labeler** – manage public/shared labels
- **administrator** – full access

Each role includes the following flags:

- **can_login** – whether accounts assigned the role are permitted to authenticate
- **is_admin** – marks administrator roles that bypass permission checks
- **public_profile_allowed_at** – when set, users with this role may expose a public profile
- **private_labels** – whether the role can use private labels

Users can hold multiple roles through the `user_roles` table. Roles are assigned
explicitly without inheritance.

The `user` role does not need to be explicitly assigned. Any authenticated
account automatically gains the `user` role while the `anyone` role applies
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
| `section`  | Permission area such as `forum`, `news`, `writing`, `imagebbs`, `images` or `role`   |
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
- **reply** – add a comment or respond in an existing thread
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
- **promote** – feature a news post as a site announcement
- **demote** – remove a post from the announcements
- **label** – create, edit or remove public/shared labels. Logged-in roles with
  view access to a section receive this grant automatically.

Sections may introduce extra actions but these form the base vocabulary used by
the templates and permission checks.
Grants with an empty `item` provide section-wide search access. For instance,
`forum|` paired with the `search` action allows a user to search all forum
topics. The grant editor uses the mapping defined in `handlers/admin/role_grants.go`
to list available actions for each section and item type. Some combinations
require an `item_id`; for example, grants in the `forum` section targeting a
`category`, `topic` or `thread` must specify the corresponding identifier.
Announcements use these actions to control which news posts appear globally. Administrator pages call `AdminPromoteAnnouncement` and `AdminDemoteAnnouncement` while `GetActiveAnnouncementWithNewsForUser` retrieves the visible announcement.

Each section may define additional actions, but these are the core verbs used by the system.

## Permission Resolution

When checking access for a user, the application resolves the user's effective
roles using the `ListEffectiveRoleIDsByUserID` query which returns the roles
directly assigned to the user.

A permission check searches the `grants` table for an active row matching the
requested `section`, `item` and `action`. A grant is considered applicable if
one of the following is true:

1. `user_id` matches the user making the request
2. `role_id` is one of the user's effective role IDs
3. both `user_id` and `role_id` are `NULL` (applies to everyone)

If no matching grant is found, the action is denied unless the user has the
`administrator` role which bypasses checks.

## Additional Tables

Some features store per-object defaults using role identifiers. Per-writing permissions are now represented using the `grants` table and no longer rely on a dedicated `writing_user_permissions` table. `imagebbs` boards are secured via grants in the `imagebbs` section using the `board` item type. Uploaded images use the `images` section with the `upload` item.

Legacy tables `topic_permissions` and `user_topic_permissions` were replaced by
equivalent rows in the `grants` table. Forum access is now fully controlled via
grants.

These tables reference roles via `role_id` columns instead of legacy `level`
fields.

## Seeding

The `seed.sql` file provides initial roles and establishes the role hierarchy.
It also grants the `user` role to existing accounts lacking a role. Migration
55 ensures all roles with `can_login` receive user-level access to the private
forum (`see`, `view`, `reply`, `post`, `edit`, `create`).

## CLI Tools

Comment-related CLI commands accept a `--user` flag to evaluate permissions for
a specific lister. The underlying queries now take a single `lister_id` rather
than separate user ID parameters.

### Default News Grants

The migrations seed baseline rules for the `news` section:

| Role | Action | Item | Description |
|------|-------|------|-------------|
| `anyone` | `see`, `view` | `post` | browse published news |
| `user` | `reply` | `post` | participate in discussions |
| `content writer`, `administrator` | `post` | `post` | create new entries |
| `content writer` | `edit` | `post` | update own news post via item-specific grant |
| `administrator` | `edit` | `post` | update any news post |

When a writer publishes a post they automatically receive an `edit` grant tied to that post, effectively granting them update rights for that item.

Other content sections such as blogs and writing follow the same pattern: authors can post entries and receive item-scoped `edit` grants while administrators hold broader `edit` privileges.
FAQ and blog listings also honour lister language preferences and check grants in SQL. Queries such as `ListBlogEntriesForLister`, `ListBlogEntriesByAuthorForLister` and `GetFAQAnsweredQuestions` filter content based on `lister_id` and permitted languages.

### Announcements

Active announcements reference a news post and are only shown to listers permitted to `view` that post. The `GetActiveAnnouncementWithNewsForLister` query filters by `lister_id` and checks the `news` section grants for the linked post.

### SQL Query Filtering

Many queries now filter results directly in SQL using `lister_id` together with the lister's effective roles. Each query matches against grants so only records the lister may access are returned. The table below lists the combinations used for each section.

| `section`  | `item`     | `action`        | Meaning |
|------------|------------|-----------------|---------|
| `blogs`    | —          | `search`        | Search blog entries |
| `blogs`    | `entry`    | `see`           | List blog entries |
| `blogs`    | `entry`    | `view`          | View a blog entry |
| `blogs`    | `entry`    | `reply`         | Comment on a blog entry |
| `blogs`    | `entry`    | `post`          | Publish a new blog entry |
| `blogs`    | `entry`    | `edit`          | Modify any blog entry |
| `faq`      | —          | `search`        | Search FAQ content |
| `faq`      | `category` | `see`           | List FAQ categories |
| `faq`      | `category` | `view`          | View questions in a FAQ category |
| `faq`      | `question` | `post`          | Submit a new FAQ question |
| `faq`      | `question/answer` | `see` | List answered FAQ questions |
| `faq`      | `question/answer` | `view` | View a FAQ question and answer |
| `forum`    | —          | `search`        | Search forums |
| `forum`    | `category` | `see`           | Discover forum categories |
| `forum`    | `category` | `view`          | View topics in the category |
| `forum`    | `category` | `post`          | Create a new topic in the category |
| `forum`    | `thread`   | `see`           | Show a thread in listings |
| `forum`    | `thread`   | `view`          | View posts within a thread |
| `forum`    | `thread`   | `reply`         | Reply within the thread |
| `forum`    | `thread`   | `edit`          | Edit posts in the thread |
| `forum`    | `topic`    | `see`           | Show a topic in listings |
| `forum`    | `topic`    | `view`          | View the topic details |
| `forum`    | `topic`    | `reply`         | Reply in the topic's threads |
| `forum`    | `topic`    | `post`          | Start a new thread in the topic |
| `forum`    | `topic`    | `edit`          | Edit threads in the topic |
| `imagebbs` | —          | `search`        | Search image boards |
| `imagebbs` | `board`    | `see`           | List image boards |
| `imagebbs` | `board`    | `view`          | View posts on a board |
| `imagebbs` | `board`    | `post`          | Create a new post on the board |
| `images`   | `upload`   | `see`           | List uploaded images |
| `images`   | `upload`   | `post`          | Upload an image |
| `privateforum` | `topic` | `see`           | Show a private topic in lists |
| `privateforum` | `topic` | `view`          | View a private topic |
| `privateforum` | `topic` | `reply`         | Reply within a private topic |
| `privateforum` | `topic` | `post`          | Start a private conversation |
| `privateforum` | `topic` | `edit`          | Edit posts in a private topic |
| `privateforum` | `topic` | `create`        | Create a private topic |
| `linker`   | —          | `search`        | Search links |
| `linker`   | `category` | `see`           | Browse link categories |
| `linker`   | `category` | `view`          | View links in a category |
| `linker`   | `category` | `post`          | Submit a link to the category |
| `linker`   | `link`     | `see`           | Show a link in lists |
| `linker`   | `link`     | `view`          | View link details |
| `linker`   | `link`     | `reply`         | Comment on a link |
| `news`     | —          | `search`        | Search news posts |
| `news`     | `post`     | `see`           | Show news posts in lists |
| `news`     | `post`     | `view`          | View a news post |
| `news`     | `post`     | `reply`         | Comment on a news post |
| `news`     | `post`     | `post`          | Publish a news post |
| `news`     | `post`     | `edit`          | Modify a news post |
| `writing`  | —          | `search`        | Search writing articles |
| `writing`  | `category` | `see`           | Browse writing categories |
| `writing`  | `category` | `view`          | View a writing category |
| `writing`  | `category` | `post`          | Publish an article in the category |
| `writing`  | `article`  | `see`           | Show writing articles in lists |
| `writing`  | `article`  | `view`          | Read a writing article |
| `writing`  | `article`  | `reply`         | Comment on a writing article |
| `writing`  | `article`  | `post`          | Publish a writing article |
| `writing`  | `article`  | `edit`          | Edit a writing article |

Viewing comments within any section uses the `view` action on the section's
primary item type since comments inherit their thread's grants and do not have
a dedicated `see` permission.

Administrator endpoints are guarded by the `AdminCheckerMiddleware` implemented
in `internal/router/router.go`. The middleware calls `corecommon.Allowed`, which
loads roles for the current user using the `GetPermissionsByUserID` query from
`internal/db/queries-permissions.sql`. Only users with the `administrator` role
can reach these routes.

Queries dealing with administration are now separated from lister paths. Queries prefixed with `Admin` such as `AdminPromoteAnnouncement` operate only in admin handlers while regular pages call the corresponding `Get...ForUser` functions like `GetActiveAnnouncementWithNewsForUser`.
Listing pages and RSS feeds still invoke `HasGrant` on each row for extra safety.

The same lookup pattern covers every action. Whether a user wants to `post`,
`reply`, `comment`, `write` or `edit`, the query matches grants by `lister_id`
and roles using the unified action names above. This ensures consistent
terminology across sections and avoids duplicated rules.
