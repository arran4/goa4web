1. **Understand the Goal**: We need to add an "Unread Threads" section in both the public and private forums. This section needs "mass actions" (i.e., a button to mark all as read). It needs to be clear and obvious.
2. **Current state**:
   - `core/templates/site/index.gohtml` already shows LHS items added via `CustomIndexItems`.
   - `handlers/forum/customindex.go` adds items to LHS for the forum.
   - `handlers/privateforum/customindex.go` does similar for the private forum.
   - I added `ListUnreadForumThreadsForUser` to `internal/db/queries-forum.sql`.
   - I added `MarkAllForumThreadsReadForUser` and `MarkAllForumThreadsNewReadForUser` to `internal/db/queries-forum.sql`.
   - I ran `go generate ./...` to generate the new queries.
3. **Execution Plan**:
   - Add LHS link "Unread Threads" to `ForumCustomIndexItems` (for `/forum` and `/private`).
   - Create `UnreadThreadsPage` handler that fetches unread threads using the new queries, reusing `ForumTopicsPageTmpl` (`forum/topicsPage.gohtml`) or creating a new template if necessary. Given the request, we can just create a `forum/unreadThreadsPage.gohtml` that includes a mass action button to mark all as read. Wait! We can reuse `topicThreads.gohtml` by creating a thin wrapper template.
   - Register routes for `/forum/unread` and `/private/unread`.
   - Create the `MarkAllUnreadReadTask` and register its route for POST to `/forum/unread/mark-all` and `/private/unread/mark-all`.
   - Verify changes with a mock test.
