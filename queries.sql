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

-- name: add_bookmarks :exec
-- This query adds a new entry to the "bookmarks" table and returns the last inserted ID as "returnthis".
INSERT INTO bookmarks (users_idusers, list)
VALUES (?, ?);
SELECT LAST_INSERT_ID() AS returnthis;

-- name: update_bookmarks :exec
-- This query updates the "list" column in the "bookmarks" table for a specific user based on their "users_idusers".
UPDATE bookmarks
SET list = ?
WHERE users_idusers = ?;

-- name: delete_bookmarks :exec
-- This query deletes all entries from the "bookmarks" table for a specific user based on their "users_idusers".
DELETE FROM bookmarks
WHERE users_idusers = ?;

-- name: show_bookmarks :one
-- This query retrieves the "list" from the "bookmarks" table for a specific user based on their "users_idusers".
SELECT list
FROM bookmarks
WHERE users_idusers = ?;

-- name: users_bookmarks :one
-- This query retrieves the "list" from the "bookmarks" table for a specific user based on their "users_idusers".
SELECT list
FROM bookmarks
WHERE users_idusers = ?;

-- name: rename_category :exec
UPDATE faqCategories
SET name = ?
WHERE idfaqCategories = ?;

-- name: delete_category :exec
DELETE FROM faqCategories
WHERE idfaqCategories = ?;

-- name: count_categories :one
SELECT COUNT(*) FROM faqCategories;

-- name: create_category :exec
INSERT INTO faqCategories (name)
VALUES (?);

-- name: add_question :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
VALUES (?, ?, ?);

-- name: reassign_category :exec
UPDATE faq
SET faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: modify_faq :exec
UPDATE faq
SET answer = ?, question = ?, faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: assign_answer :exec
UPDATE faq
SET answer = ?
WHERE idfaq = ?;

-- name: categories :one
SELECT idfaqCategories, name
FROM faqCategories;

-- name: category_faqs :many
SELECT question, idfaq, answer, faqCategories_idfaqCategories
FROM faq
WHERE faqCategories_idfaqCategories = ? OR answer IS NULL;

-- name: show_questions :many
SELECT c.idfaqCategories, c.name, f.question, f.answer
FROM faq f, faqCategories c
WHERE c.idfaqCategories <> ? AND f.answer IS NOT NULL AND c.idfaqCategories = f.faqCategories_idfaqCategories AND (c.idfaqCategories = ?)
ORDER BY c.idfaqCategories;

-- name: admin_categories :many
SELECT idfaqCategories, name
FROM faqCategories;

-- name: show_categories :exec
SELECT f.idforumcategory, f.title, f.description
FROM forumcategory f WHERE f.forumcategory_idforumcategory = $1;

-- name: changeCategory :exec
UPDATE forumcategory SET title = $2, description = $3 WHERE idforumcategory = $1;

-- name: showAllCategories :many
SELECT c.idforumcategory, c.title, c.description, c.forumcategory_idforumcategory, c2.title
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.forumcategory_idforumcategory = c2.idforumcategory;

-- name: showAllTopics :many
SELECT t.idforumtopic, t.title, t.description, t.forumcategory_idforumcategory, c.title
FROM forumtopic t
LEFT JOIN forumcategory c ON t.forumcategory_idforumcategory = c.idforumcategory
GROUP BY t.idforumtopic;

-- name: changeTopic :exec
UPDATE forumtopic SET title = $2, description = $3 WHERE idforumtopic = $1;

-- name: show_topics :many
SELECT t.idforumtopic, t.title, t.description, t.comments, t.threads, t.lastaddition, lu.username, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = $1
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = $2
ORDER BY t.lastaddition DESC;

-- name: printTopic :many
SELECT LEFT(c.text, 255), fu.username, c.written, lu.username, t.lastaddition, t.idforumthread, t.comments, r.viewlevel, u.level
FROM forumthread t
LEFT JOIN topicrestrictions r ON t.forumtopic_idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.forumtopic_idforumtopic AND u.users_idusers = $1
LEFT JOIN comments c ON c.idcomments = t.firstpost
LEFT JOIN users fu ON fu.idusers = c.users_idusers
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumtopic_idforumcategory = $2
ORDER BY t.lastaddition DESC;

-- name: deleteTopicRestrictions :exec
DELETE FROM topicrestrictions WHERE forumtopic_idforumtopic = $1;

-- name: existsTopicRestrictions :one
SELECT (forumtopic_idforumtopic) FROM topicrestrictions WHERE forumtopic_idforumtopic = $1;

-- name: addTopicRestrictions :exec
INSERT INTO topicrestrictions (forumtopic_idforumtopic, viewlevel, replylevel, newthreadlevel, seelevel, invitelevel, readlevel, modlevel, adminlevel)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: setTopicRestrictions :exec
UPDATE topicrestrictions SET viewlevel = $1, replylevel = $2, newthreadlevel = $3, seelevel = $4, invitelevel = $5, readlevel = $6, modlevel = $7, adminlevel = $8
WHERE forumtopic_idforumtopic = $9;

-- name: printTopicRestrictions :many
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = $1;

-- name: deleteUsersTopicLevel :exec
DELETE FROM userstopiclevel WHERE forumtopic_idforumtopic = $1 AND users_idusers = $2;

-- name: addUsersTopicLevel :exec
INSERT INTO userstopiclevel (forumtopic_idforumtopic, users_idusers, level, invitemax)
VALUES ($1, $2, $3, $4);

-- name: setUsersTopicLevel :exec
UPDATE userstopiclevel SET level = $3, invitemax = $4 WHERE forumtopic_idforumtopic = $1 AND users_idusers = $2;

-- name: getUsersTopicLevelInviteMax :one
SELECT invitemax FROM userstopiclevel WHERE forumtopic_idforumtopic = $1 AND users_idusers = $2;

-- name: getUsersTopicLevel :one
SELECT level FROM userstopiclevel WHERE forumtopic_idforumtopic = $1 AND users_idusers = $2;

-- name: showTopicUserLevels :one
SELECT r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = $1;

-- name: showTableTopics :many
SELECT t.idforumtopic, t.title, t.description, t.comments, t.threads, t.lastaddition, lu.username, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = $1
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE forumcategory_idforumcategory = $2 AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: expandCategories :many
SELECT f.idforumcategory, f.title, f.description
FROM forumcategory f WHERE f.forumcategory_idforumcategory = $1;

-- name: printCategoryRoots :many
SELECT c3.idforumcategory, c3.title, c2.idforumcategory, c2.title, c1.title
FROM forumcategory c1
LEFT JOIN forumcategory c2 ON c2.idforumcategory = c1.forumcategory_idforumcategory
LEFT JOIN forumcategory c3 ON c3.idforumcategory = c2.forumcategory_idforumcategory
WHERE c1.idforumcategory = $1;

-- name: printTopicRoots :many
SELECT c3.idforumcategory, c3.title, c2.idforumcategory, c2.title, c1.idforumcategory, c1.title, t.title
FROM forumtopic t
LEFT JOIN forumcategory c1 ON c1.idforumcategory = t.forumcategory_idforumcategory
LEFT JOIN forumcategory c2 ON c2.idforumcategory = c1.forumcategory_idforumcategory
LEFT JOIN forumcategory c3 ON c3.idforumcategory = c2.forumcategory_idforumcategory
WHERE t.idforumtopic = $1;
