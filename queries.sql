-- name: renameLanguage :exec
-- This query updates the "nameof" field in the "language" table based on the provided "cid".
-- Parameters:
--   $1 - New name for the language (string)
--   $2 - Language ID to be updated (int)
UPDATE language
SET nameof = $1
WHERE idlanguage = $2;

-- name: deleteLanguage :exec
-- This query deletes a record from the "language" table based on the provided "cid".
-- Parameters:
--   $1 - Language ID to be deleted (int)
DELETE FROM language
WHERE idlanguage = $1;

-- name: countCategories :one
-- This query returns the count of all records in the "language" table.
-- Result:
--   count(*) - The count of rows in the "language" table (int)
SELECT COUNT(*) AS count
FROM language;

-- name: createLanguage :exec
-- This query inserts a new record into the "language" table.
-- Parameters:
--   $1 - Name of the new language (string)
INSERT INTO language (nameof)
VALUES ($1);

-- name: SelectLanguages :many
-- This query selects all languages from the "language" table.
-- Result:
--   idlanguage (int)
--   nameof (string)
SELECT idlanguage, nameof
FROM language;

-- name: adminUserPermissions :many
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT p.idpermissions, p.level, u.username, u.email, p.section
FROM permissions p, users u
WHERE u.idusers = p.users_idusers
ORDER BY p.level;

-- name: userAllow :exec
-- This query inserts a new permission into the "permissions" table.
-- Parameters:
--   $1 - User ID to be associated with the permission (int)
--   $2 - Section for which the permission is granted (string)
--   $3 - Level of the permission (string)
INSERT INTO permissions (users_idusers, section, level)
VALUES ($1, $2, $3);

-- name: userDisallow :exec
-- This query deletes a permission from the "permissions" table based on the provided "permid".
-- Parameters:
--   $1 - Permission ID to be deleted (int)
DELETE FROM permissions
WHERE idpermissions = $1;

-- name: adminUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.idusers, u.username, u.email
FROM users u;

-- name: completeWordList :exec
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: remakeCommentsSearch :exec
-- This query selects data from the "comments" table and populates the "commentsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "commentsSearch" using the "comments_idcomments".
INSERT INTO commentsSearch (text, comments_idcomments)
SELECT text, idcomments
FROM comments;
DELETE FROM commentsSearch;

-- name: remakeNewsSearch :exec
-- This query selects data from the "siteNews" table and populates the "siteNewsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "siteNews_idsiteNews".
INSERT INTO siteNewsSearch (text, siteNews_idsiteNews)
SELECT news, idsiteNews
FROM siteNews;
DELETE FROM siteNewsSearch;

-- name: remakeBlogSearch :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blogs_idblogs".
INSERT INTO blogsSearch (text, blogs_idblogs)
SELECT blog, idblogs
FROM blogs;
DELETE FROM blogsSearch;

-- name: remakeWritingSearch :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_idwriting".
INSERT INTO writingSearch (text, writing_idwriting)
SELECT CONCAT(title, ' ', abstract, ' ', writting), idwriting
FROM writing;
DELETE FROM writingSearch;

-- name: remakeLinkerSearch :exec
-- This query selects data from the "linker" table and populates the "linkerSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linkerSearch" using the "linker_idlinker".
INSERT INTO linkerSearch (text, linker_idlinker)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;
DELETE FROM linkerSearch;

-- name: remakeCommentsSearchInsert :exec
-- This query selects data from the "comments" table and populates the "commentsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "commentsSearch" using the "comments_idcomments".
INSERT INTO commentsSearch (text, comments_idcomments)
SELECT text, idcomments
FROM comments;

-- name: deleteCommentsSearch :exec
-- This query deletes all data from the "commentsSearch" table.
DELETE FROM commentsSearch;

-- name: remakeNewsSearchInsert :exec
-- This query selects data from the "siteNews" table and populates the "siteNewsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "siteNews_idsiteNews".
INSERT INTO siteNewsSearch (text, siteNews_idsiteNews)
SELECT news, idsiteNews
FROM siteNews;

-- name: deleteSiteNewsSearch :exec
-- This query deletes all data from the "siteNewsSearch" table.
DELETE FROM siteNewsSearch;

-- name: remakeBlogsSearchInsert :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blogs_idblogs".
INSERT INTO blogsSearch (text, blogs_idblogs)
SELECT blog, idblogs
FROM blogs;

-- name: deleteBlogsSearch :exec
-- This query deletes all data from the "blogsSearch" table.
DELETE FROM blogsSearch;

-- name: remakeWritingSearchInsert :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_idwriting".
INSERT INTO writingSearch (text, writing_idwriting)
SELECT CONCAT(title, ' ', abstract, ' ', writting), idwriting
FROM writing;

-- name: deleteWritingSearch :exec
-- This query deletes all data from the "writingSearch" table.
DELETE FROM writingSearch;

-- name: remakeLinkerSearchInsert :exec
-- This query selects data from the "linker" table and populates the "linkerSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linkerSearch" using the "linker_idlinker".
INSERT INTO linkerSearch (text, linker_idlinker)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;

-- name: deleteLinkerSearch :exec
-- This query deletes all data from the "linkerSearch" table.
DELETE FROM linkerSearch;

-- name: update_forumthread_lastaddition :exec
-- This query updates the "lastaddition" column in the "forumthread" table.
-- It sets the "lastaddition" column to the latest "written" value from the "comments" table for the corresponding "forumthread_idforumthread".
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
);

-- name: update_forumthread_comments :exec
-- This query updates the "comments" column in the "forumthread" table.
-- It sets the "comments" column to the count of users (excluding the thread creator) from the "comments" table for the corresponding "forumthread_idforumthread".
UPDATE forumthread
SET comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
);

-- name: update_forumthread_lastposter :exec
-- This query updates the "lastposter" column in the "forumthread" table.
-- It sets the "lastposter" column to the latest "users_idusers" value from the "comments" table for the corresponding "forumthread_idforumthread".
UPDATE forumthread
SET lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
);

-- name: update_forumthread_firstpost :exec
-- This query updates the "firstpost" column in the "forumthread" table.
-- It sets the "firstpost" column to the ID of the first comment from the "comments" table for the corresponding "forumthread_idforumthread".
UPDATE forumthread
SET firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    LIMIT 1
);

-- name: update_forumtopic_threads :exec
-- This query updates the "threads" column in the "forumtopic" table.
-- It sets the "threads" column to the count of forum threads from the "forumthread" table for the corresponding "forumtopic_idforumtopic".
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
);

-- name: update_forumtopic_comments :exec
-- This query updates the "comments" column in the "forumtopic" table.
-- It sets the "comments" column to the sum of comments from the "forumthread" table for the corresponding "forumtopic_idforumtopic".
UPDATE forumtopic
SET comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
);

-- name: update_forumtopic_lastaddition_lastposter :exec
-- This query updates the "lastaddition" and "lastposter" columns in the "forumtopic" table.
-- It sets the "lastaddition" column to the latest "lastaddition" value from the "forumthread" table for the corresponding "forumtopic_idforumtopic".
-- It sets the "lastposter" column to the latest "lastposter" value from the "forumthread" table for the corresponding "forumtopic_idforumtopic".
UPDATE forumtopic
SET lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
),
lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
);

-- name: blogid_to_userid :one
SELECT idusers
FROM users u, blogs b
WHERE u.idusers = b.users_idusers AND b.idblogs = ?;

-- name: delete_blog :exec
DELETE FROM blogs
WHERE idblogs = ?;

-- name: delete_blog_search :exec
DELETE FROM blogsSearch
WHERE blogs_idblogs = ?;

-- name: update_blog :exec
UPDATE blogs
SET language_idlanguage = ?, blog = ?
WHERE idblogs = ?;

-- name: delete_blog_comments :exec
DELETE FROM comments
WHERE forumthread_idforumthread = ?;

-- name: add_blog :exec
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
VALUES (?, ?, ?, NOW());
SELECT LAST_INSERT_ID() AS value;

-- name: assign_blog_to_thread :exec
UPDATE blogs
SET forumthread_idforumthread = ?
WHERE idblogs = ?;

-- name: show_latest_blogs :many
SELECT b.blog, b.written, u.username, b.idblogs, IF(th.comments IS NULL, 0, th.comments + 1), b.users_idusers
FROM blogs b, users u
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.users_idusers = ? AND (b.language_idlanguage = ?)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: show_blog_comments :many
SELECT b.blog, b.written, u.username, b.idblogs, b.forumthread_idforumthread
FROM blogs b, users u
WHERE b.users_idusers = u.idusers AND b.idblogs = ?;

-- name: show_blogger_list :many
SELECT u.username, COUNT(b.idblogs)
FROM blogs b, users u
WHERE b.users_idusers = u.idusers
GROUP BY u.idusers;

-- name: write_blog_atom :many
SELECT b.idblogs, LEFT(b.written, 255), b.blog, u.username
FROM blogs b, users u
WHERE u.idusers = b.users_idusers AND b.users_idusers = ?
ORDER BY b.written DESC
LIMIT ?;

-- name: write_blog_rss :many
SELECT b.idblogs, LEFT(b.written, 255), b.blog, u.username
FROM blogs b, users u
WHERE u.idusers = b.users_idusers AND b.users_idusers= ?
ORDER BY b.written DESC
LIMIT ?;

-- name: admin_user_permissions :many
SELECT p.idpermissions, p.level, u.username, u.email, p.section
FROM permissions p, users u
WHERE u.idusers = p.users_idusers AND p.section = ?
ORDER BY p.level;

-- name: user_allow :exec
INSERT INTO permissions (users_idusers, section, level)
VALUES (?, ?, ?);

-- name: user_disallow :exec
DELETE FROM permissions
WHERE idpermissions = ? AND section = ?;

-- name: show_blog_edit :many
SELECT b.blog, b.language_idlanguage
FROM blogs b, users u
WHERE b.users_idusers = u.idusers AND b.idblogs = ?;
