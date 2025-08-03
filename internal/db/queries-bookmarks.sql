-- name: CreateBookmarksForLister :exec
-- This query adds a new entry to the "bookmarks" table for a lister.
INSERT INTO bookmarks (users_idusers, list)
VALUES (?, ?);

-- name: UpdateBookmarksForLister :exec
-- This query updates the "list" column in the "bookmarks" table for a specific lister.
UPDATE bookmarks b
SET list = sqlc.arg(list)
WHERE b.users_idusers = sqlc.arg(lister_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='bookmarks'
        AND (g.item='list' OR g.item IS NULL)
        AND g.action='post'
        AND g.active=1
        AND (g.item_id = 0 OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
        ))
  );

-- name: GetBookmarksForUser :one
-- This query retrieves the "list" from the "bookmarks" table for a specific user based on their "users_idusers".
SELECT Idbookmarks, list
FROM bookmarks
WHERE users_idusers = ?;

