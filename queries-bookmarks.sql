-- name: Add_bookmarks :exec
-- This query adds a new entry to the "bookmarks" table and returns the last inserted ID as "returnthis".
INSERT INTO bookmarks (users_idusers, list)
VALUES (?, ?);
SELECT LAST_INSERT_ID() AS returnthis;

-- name: Update_bookmarks :exec
-- This query updates the "list" column in the "bookmarks" table for a specific user based on their "users_idusers".
UPDATE bookmarks
SET list = ?
WHERE users_idusers = ?;

-- name: Show_bookmarks :one
-- This query retrieves the "list" from the "bookmarks" table for a specific user based on their "users_idusers".
SELECT Idbookmarks, list
FROM bookmarks
WHERE users_idusers = ?;

