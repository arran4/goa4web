# SQL Query Naming Update Tasks

The following tasks continue the migration towards explicit query names that
indicate whether an operation is for a user, an administrator, or a system
background task. Each task touches only a single package to minimise merge
conflicts.

1. **Announcements** – rename `ListAnnouncementsWithNews` to
   `AdminListAnnouncementsWithNews` and change `CreateAnnouncement`/
   `DeleteAnnouncement` to `PromoteAnnouncement`/`DemoteAnnouncement`.
   Add language filtering and rename `GetActiveAnnouncementWithNews` to
   `GetActiveAnnouncementWithNewsForUser`.
2. **Audit Log** – prefix all queries in `queries-auditlog.sql` with `Admin` and
   adjust call sites accordingly.
3. **Banned IPs** – prefix queries in `queries-banned_ips.sql` with `Admin`.
4. **Blog Entry Grants** – add grant checks to
   `CreateBlogEntry` and `UpdateBlogEntry`. Rename
   `GetBlogEntriesForUserDescending` to
   `GetBlogEntriesDescendingForViewer` and ensure language filtering.
5. **Blog Search** – introduce `BlogsSearchFirstForViewer` and
   `BlogsSearchNextForViewer` with grant checks.
6. **Blog Listing** – rename `GetAllBlogEntriesByUser` to
   `AdminGetAllBlogEntriesByUser`.
7. **Bloggers** – verify the queries and rename them to
   `ListBloggersForViewer` and `SearchBloggersForViewer` if not already.
8. **Blog Index Maintenance** – mark `SetBlogLastIndex` and
   `GetAllBlogsForIndex` as system queries.
9. **Comments** – add grant checks to `GetCommentsByThreadIdForUser` and rename
   `GetAllCommentsByUser` and `ListAllCommentsWithThreadInfo` with `Admin`
   prefixes.
10. **Deactivation, DLQ and External Links** – prefix all queries in
    `queries-deactivation.sql`, `queries-dlq.sql` and
    `queries-externallinks.sql` with `Admin` or `System` as appropriate.
11. **FAQ Management** – rename admin queries such as
    `GetFAQUnansweredQuestions`, `GetFAQAnsweredQuestions`,
    `GetFAQDismissedQuestions` and others using the `Admin` prefix.
    Introduce `GetAllFAQQuestionsForViewer` with language and grant
    filtering and update related call sites.
12. **FAQ Revisions** – split `InsertFAQRevision` into a user-facing variant
    with grant checks and an admin variant for manual edits.
13. **Role and Permission Queries** – audit remaining SQL files and apply the
    same naming scheme so each query clearly states whether it is for a user,
    administrator or system task.

These tasks provide a roadmap for aligning the query layer with the naming
convention described in `permissions.md`.
