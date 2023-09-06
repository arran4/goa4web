-- name: renameLanguage :exec
-- This query updates the "nameof" field in the "language" table based on the provided "cid".
-- Parameters:
--   ? - New name for the language (string)
--   ? - Language ID to be updated (int)
UPDATE language
SET nameof = ?
WHERE idlanguage = ?;

-- name: deleteLanguage :exec
-- This query deletes a record from the "language" table based on the provided "cid".
-- Parameters:
--   ? - Language ID to be deleted (int)
DELETE FROM language
WHERE idlanguage = ?;

-- name: countCategories :one
-- This query returns the count of all records in the "language" table.
-- Result:
--   count(*) - The count of rows in the "language" table (int)
SELECT COUNT(*) AS count
FROM language;

-- name: createLanguage :exec
-- This query inserts a new record into the "language" table.
-- Parameters:
--   ? - Name of the new language (string)
INSERT INTO language (nameof)
VALUES (?);

-- name: SelectLanguages :many
-- This query selects all languages from the "language" table.
-- Result:
--   idlanguage (int)
--   nameof (string)
SELECT idlanguage, nameof
FROM language;

-- name: getUserPermissions :one
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT p.*
FROM permissions p
WHERE p.users_idusers = ?
;

-- name: getUsersPermissions :many
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT p.*
FROM permissions p
;

-- name: getUsersTopicLevel :one
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT utl.*
FROM userstopiclevel utl
WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?
;

-- name: userAllow :exec
-- This query inserts a new permission into the "permissions" table.
-- Parameters:
--   ? - User ID to be associated with the permission (int)
--   ? - Section for which the permission is granted (string)
--   ? - Level of the permission (string)
INSERT INTO permissions (users_idusers, section, level)
VALUES (?, ?, ?);

-- name: userDisallow :exec
-- This query deletes a permission from the "permissions" table based on the provided "permid".
-- Parameters:
--   ? - Permission ID to be deleted (int)
DELETE FROM permissions
WHERE idpermissions = ?;

-- name: allUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.*
FROM users u;

-- name: completeWordList :many
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: remakeCommentsSearch :exec
-- This query selects data from the "comments" table and populates the "commentsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "commentsSearch" using the "comments_idcomments".
INSERT INTO commentsSearch (text, comments_idcomments)
SELECT text, idcomments
FROM comments;

-- name: remakeNewsSearch :exec
-- This query selects data from the "siteNews" table and populates the "siteNewsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "siteNews_idsiteNews".
INSERT INTO siteNewsSearch (text, siteNews_idsiteNews)
SELECT news, idsiteNews
FROM siteNews;

-- name: remakeBlogSearch :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blogs_idblogs".
INSERT INTO blogsSearch (text, blogs_idblogs)
SELECT blog, idblogs
FROM blogs;

-- name: remakeWritingSearch :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_idwriting".
INSERT INTO writingSearch (text, writing_idwriting)
SELECT CONCAT(title, ' ', abstract, ' ', writting), idwriting
FROM writing;

-- name: remakeLinkerSearch :exec
-- This query selects data from the "linker" table and populates the "linkerSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linkerSearch" using the "linker_idlinker".
INSERT INTO linkerSearch (text, linker_idlinker)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;

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

-- name: update_forumthreads :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
), lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    LIMIT 1
);

-- name: update_forumtopics :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
), lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
);

-- name: update_forumthread :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
), lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    LIMIT 1
)
WHERE idforumthread = ?;

-- name: update_forumtopic :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
), lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
)
WHERE idforumtopic = ?;

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

-- name: update_comment :exec
UPDATE comments
SET language_idlanguage = ?, text = ?
WHERE idcomments = ?;

-- name: delete_blog_comments :exec
DELETE FROM comments
WHERE forumthread_idforumthread = ?;

-- name: add_blog :execlastid
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
VALUES (?, ?, ?, NOW());

-- name: assign_blog_to_thread :exec
UPDATE blogs
SET forumthread_idforumthread = ?
WHERE idblogs = ?;

-- name: show_latest_blogs :many
SELECT b.blog, b.written, u.username, b.idblogs, coalesce(th.comments, 0), b.users_idusers
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: getBlogs :many
SELECT b.*
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
-- WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
-- AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
-- AND b.idblogs IN (sqlc.slice(blogIds))
WHERE b.idblogs IN (sqlc.slice(blogIds))
ORDER BY b.written DESC
-- LIMIT ? OFFSET ?
;

-- name: show_blog :one
SELECT b.blog, b.written, u.username, b.idblogs, coalesce(th.comments, 0), b.users_idusers, b.forumthread_idforumthread
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.idblogs = ?
LIMIT 1;

-- name: user_get_comment :one
SELECT c.*, pu.Username
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = ? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
LIMIT 1;

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

-- name: show_blog_edit :one
SELECT b.blog, b.language_idlanguage
FROM blogs b, users u
WHERE b.users_idusers = u.idusers AND b.idblogs = ?
LIMIT 1;

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
SELECT Idbookmarks, list
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

-- name: delete_faq :exec
DELETE FROM faq
WHERE idfaq = ?;

-- name: faq_categories :many
SELECT idfaqCategories, name
FROM faqCategories;

-- name: category_faqs :many
SELECT question, idfaq, answer, faqCategories_idfaqCategories
FROM faq
WHERE faqCategories_idfaqCategories = ? OR answer IS NULL;

-- name: show_questions :many
SELECT c.idfaqCategories, c.name, f.question, f.answer
FROM faq f
LEFT JOIN faqCategories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0 AND f.answer IS NOT NULL
ORDER BY c.idfaqCategories;

-- name: admin_categories :many
SELECT idfaqCategories, name
FROM faqCategories;

-- name: show_categories :exec
SELECT f.idforumcategory, f.title, f.description
FROM forumcategory f WHERE f.forumcategory_idforumcategory = ?;

-- name: changeCategory :exec
UPDATE forumcategory SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumcategory = ?;

-- name: showAllCategories :many
SELECT c.*, COUNT(c2.idforumcategory) as SubcategoryCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.forumcategory_idforumcategory = c2.idforumcategory
GROUP BY c.idforumcategory;

-- name: getAllTopics :many
SELECT t.*
FROM forumtopic t
LEFT JOIN forumcategory c ON t.forumcategory_idforumcategory = c.idforumcategory
GROUP BY t.idforumtopic;

-- name: changeTopic :exec
UPDATE forumtopic SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumtopic = ?;

-- name: get_all_user_topics_for_category :many
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = ? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: get_all_user_topics :many
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: user_get_topic :one
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0) AND t.idforumtopic=?
ORDER BY t.lastaddition DESC;

-- name: user_get_thread :one
SELECT th.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0) AND th.idforumthread=?
ORDER BY t.lastaddition DESC;

-- name: deleteTopicRestrictions :exec
DELETE FROM topicrestrictions WHERE forumtopic_idforumtopic = ?;

-- name: existsTopicRestrictions :one
SELECT (forumtopic_idforumtopic) FROM topicrestrictions WHERE forumtopic_idforumtopic = ?;

-- name: addTopicRestrictions :exec
INSERT INTO topicrestrictions (forumtopic_idforumtopic, viewlevel, replylevel, newthreadlevel, seelevel, invitelevel, readlevel, modlevel, adminlevel)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: setTopicRestrictions :exec
INSERT INTO topicrestrictions (forumtopic_idforumtopic, viewlevel, replylevel, newthreadlevel, seelevel, invitelevel, readlevel, modlevel, adminlevel)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    viewlevel = VALUES(viewlevel),
    replylevel = VALUES(replylevel),
    newthreadlevel = VALUES(newthreadlevel),
    seelevel = VALUES(seelevel),
    invitelevel = VALUES(invitelevel),
    readlevel = VALUES(readlevel),
    modlevel = VALUES(modlevel),
    adminlevel = VALUES(adminlevel);

-- name: getTopicRestrictions :many
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: getAllTopicRestrictions :many
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic;

-- name: deleteUsersTopicLevel :exec
DELETE FROM userstopiclevel WHERE forumtopic_idforumtopic = ? AND users_idusers = ?;

-- name: addUsersTopicLevel :exec
INSERT INTO userstopiclevel (forumtopic_idforumtopic, users_idusers, level, invitemax)
VALUES (?, ?, ?, ?);

-- name: setUsersTopicLevel :exec
INSERT INTO userstopiclevel (forumtopic_idforumtopic, users_idusers, level, invitemax)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE level = VALUES(level), invitemax = VALUES(invitemax);

-- name: getAllUsersTopicLevelInviteMax :one
SELECT invitemax FROM userstopiclevel WHERE forumtopic_idforumtopic = ? AND users_idusers = ?;

-- name: getUsersAllTopicLevels :many
SELECT u.*, t.*, utl.*, tr.*
FROM users u
JOIN userstopiclevel utl ON utl.users_idusers=u.idusers
JOIN forumtopic t ON utl.forumtopic_idforumtopic = t.idforumtopic
JOIN topicrestrictions tr ON t.idforumtopic = tr.forumtopic_idforumtopic
WHERE u.idusers = ?;

-- name: getAllUsersAllTopicLevels :many
SELECT u.*, t.*, utl.*, tr.*
FROM users u
JOIN userstopiclevel utl ON utl.users_idusers=u.idusers
JOIN forumtopic t ON utl.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN topicrestrictions tr ON t.idforumtopic = tr.forumtopic_idforumtopic;

-- name: showTopicUserLevels :one
SELECT r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: forumCategories :many
SELECT f.*
FROM forumcategory f;

-- name: writeRSS :exec
SELECT title, description FROM imageboard WHERE idimageboard = ?;

-- name: makeImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description) VALUES (?, ?, ?);

-- name: changeImageBoard :exec
UPDATE imageboard SET title = ?, description = ?, imageboard_idimageboard = ? WHERE idimageboard = ?;

-- name: printSubBoards :many
SELECT idimageboard, title, description FROM imageboard WHERE imageboard_idimageboard = ?;

-- name: printImagePosts :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.imageboard_idimageboard = ?;

-- name: printImagePost :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.idimagepost = ?;

-- name: printBoardPosts :many
SELECT i.description, i.thumbnail, i.fullimage, u.username, i.posted, i.idimagepost, IF(th.comments IS NULL, 0, th.comments + 1)
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.imageboard_idimageboard = ?
ORDER BY i.posted DESC;

-- name: addImage :exec
INSERT INTO imagepost (imageboard_idimageboard, thumbnail, fullimage, users_idusers, description, posted)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: assignImagePostThisThreadId :exec
UPDATE imagepost SET forumthread_idforumthread = ? WHERE idimagepost = ?;

-- name: showAllBoards :many
SELECT b.idimageboard, b.title, b.description, b.imageboard_idimageboard, pb.title
FROM imageboard b
LEFT JOIN imageboard pb ON b.imageboard_idimageboard = pb.idimageboard OR b.imageboard_idimageboard = 0
GROUP BY b.idimageboard;

-- name: writeSiteNewsRSS :many
SELECT s.idsiteNews, s.occured, s.news
FROM siteNews s
ORDER BY s.occured DESC LIMIT 15;

-- name: writeNewsPost :exec
INSERT INTO siteNews (news, users_idusers, occured, language_idlanguage)
VALUES (?, ?, NOW(), ?);

-- name: editNewsPost :exec
UPDATE siteNews SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?;

-- name: doCalled :many
SELECT s.news, s.idsiteNews, u.idusers, s.language_idlanguage
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: getNewsThreadId :one
SELECT s.forumthread_idforumthread, u.idusers
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: assignNewsThisThreadId :exec
UPDATE siteNews SET forumthread_idforumthread = ? WHERE idsiteNews = ?;

-- name: getNewsPost :one
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.idsiteNews = ?
;

-- name: getNewsPosts :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds))
;

-- name: getLatestNewsPosts :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
ORDER BY s.occured DESC
LIMIT 15
;

-- -- name: showNews :one
-- SELECT count(idsiteNews) FROM siteNews
-- WHERE ? AND ?;

-- -- name: showNewsPosts :many
-- SELECT u.username, s.news, s.occured, s.idsiteNews, u.idusers, IF(th.comments IS NULL, 0, th.comments + 1)
-- FROM siteNews s
-- LEFT JOIN users u ON s.users_idusers = u.idusers
-- LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
-- WHERE ? AND ?
-- ORDER BY s.occured DESC
-- LIMIT 10;

-- name: countLinkerCategories :one
SELECT COUNT(*) FROM linkerCategory;

-- name: deleteCategory :exec
DELETE FROM linkerCategory WHERE idlinkerCategory = ?;

-- name: renameCategory :exec
UPDATE linkerCategory SET title = ? WHERE idlinkerCategory = ?;

-- name: createCategory :exec
INSERT INTO linkerCategory (title) VALUES (?);

-- name: showCategories :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: category_combobox :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: adminCategories :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: deleteQueueItem :exec
DELETE FROM linkerQueue WHERE idlinkerQueue = ?;

-- name: updateQueue :exec
UPDATE linkerQueue SET linkerCategory_idlinkerCategory = ?, title = ?, url = ?, description = ? WHERE idlinkerQueue = ?;

-- name: addToQueue :exec
INSERT INTO linkerQueue (users_idusers, linkerCategory_idlinkerCategory, title, url, description) VALUES (?, ?, ?, ?, ?);

-- name: showAdminQueue :many
SELECT l.*, u.username, c.title as category_title, c.idlinkerCategory
FROM linkerQueue l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory c ON l.linkerCategory_idlinkerCategory = c.idlinkerCategory
;
-- name: moveToLinker :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, language_idlanguage, title, `url`, description)
SELECT l.users_idusers, l.linkerCategory_idlinkerCategory, l.language_idlanguage, l.title, l.url, l.description
FROM linkerQueue l
WHERE l.idlinkerQueue = ?
;

-- name: addToLinker :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, title, url, description, listed)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: assignLinkerThisThreadId :exec
UPDATE linker SET forumthread_idforumthread = ? WHERE idlinker = ?;

-- name: showLatest :many
SELECT l.*, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_idforumthread = th.idforumthread
WHERE lc.idlinkerCategory = ?
ORDER BY l.listed DESC;

-- name: showLink :one
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker = ?;

-- name: showLinks :many
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds));

-- name: writeLinkerRSS :many
SELECT l.idlinker, l.title, l.description, l.url
FROM linker l
WHERE l.linkerCategory_idlinkerCategory = ?
ORDER BY l.listed DESC;

-- -- name: forumTopicSearch :many
-- SELECT * FROM comments c
-- LEFT JOIN forumthread th ON th.idforumthread = c.forumthread_idforumthread
-- LEFT JOIN forumtopic t ON t.idforumtopic = th.forumtopic_idforumtopic
-- LEFT JOIN userstopiclevel utl ON t.idforumtopic = utl.forumtopic_idforumtopic AND utl.users_idusers = ?
-- LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
--  WHERE c.idcomments IN (?) AND th.idforumthread != 0 AND t.idforumtopic = ?
-- AND ((r.readlevel <= utl.level AND r.viewlevel <= utl.level AND r.seelevel <= utl.level));
--
-- -- name: forumSearch :many
-- SELECT c.forumthread_idforumthread FROM comments c
-- LEFT JOIN forumthread th ON th.idforumthread = c.forumthread_idforumthread
-- LEFT JOIN forumtopic t ON t.idforumtopic = th.forumtopic_idforumtopic
-- LEFT JOIN forumcategory fc ON fc.idforumcategory = t.forumcategory_idforumcategory
-- LEFT JOIN userstopiclevel utl ON t.idforumtopic = utl.forumtopic_idforumtopic AND utl.users_idusers = ?
-- LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
-- WHERE c.idcomments IN (?) AND th.idforumthread != 0 AND t.idforumtopic != 0
-- AND ((r.readlevel <= utl.level AND r.viewlevel <= utl.level AND r.seelevel <= utl.level) OR ?)
-- AND fc.idforumcategory != 0
-- GROUP BY c.forumthread_idforumthread;

-- name: usernametouid :one
SELECT idusers FROM users WHERE username = ?;

-- name: lang_combobox :many
SELECT l.idlanguage, l.nameof FROM language l;

-- name: getSecurityLevel :one
SELECT level FROM permissions WHERE users_idusers = ? AND (section = ? OR section = 'all');

-- name: getLangs :one
SELECT language_idlanguage FROM userlang WHERE users_idusers = ?;

-- name: preferencesRefreshPref :many
SELECT language_idlanguage FROM preferences WHERE users_idusers = ?;

-- name: getWordID :one
SELECT idsearchwordlist FROM searchwordlist WHERE word = lcase(?);

-- name: addWord :execlastid
INSERT IGNORE INTO searchwordlist (word)
VALUES (lcase(sqlc.arg(word)));

-- name: addToForumCommentSearch :exec
INSERT IGNORE INTO commentsSearch
(comments_idcomments, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: commentsSearchDelete :exec
DELETE FROM commentsSearch
WHERE comments_idcomments=?
;

-- name: commentsSearchFirstNotInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND ft.forumcategory_idforumcategory!=0
;

-- name: commentsSearchNextNotInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND cs.comments_idcomments IN (sqlc.slice('ids'))
AND ft.forumcategory_idforumcategory!=0
;

-- name: commentsSearchFirstInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
WHERE swl.word=?
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: commentsSearchNextInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
WHERE swl.word=?
AND cs.comments_idcomments IN (sqlc.slice('ids'))
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: addToForumWritingSearch :exec
INSERT IGNORE INTO writingSearch
(writing_idwriting, searchwordlist_idsearchwordlist)
VALUES (?, ?);


-- name: writingSearchDelete :exec
DELETE FROM writingSearch
WHERE writing_idwriting=?
;

-- name: writingSearchFirst :many
SELECT DISTINCT cs.writing_idwriting
FROM writingSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: writingSearchNext :many
SELECT DISTINCT cs.writing_idwriting
FROM writingSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.writing_idwriting IN (sqlc.slice('ids'))
;

-- name: addToForumSiteNewsearch :exec
INSERT IGNORE INTO siteNewsSearch
(siteNews_idsiteNews, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: siteNewsSearchDelete :exec
DELETE FROM siteNewsSearch
WHERE siteNews_idsiteNews=?
;

-- name: siteNewsSearchFirst :many
SELECT DISTINCT cs.siteNews_idsiteNews
FROM siteNewsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: siteNewsSearchNext :many
SELECT DISTINCT cs.siteNews_idsiteNews
FROM siteNewsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.siteNews_idsiteNews IN (sqlc.slice('ids'))
;

-- name: linkerSearchDelete :exec
DELETE FROM linkerSearch
WHERE linker_idlinker=?
;

-- name: linkerSearchFirst :many
SELECT DISTINCT cs.linker_idlinker
FROM linkerSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: linkerSearchNext :many
SELECT DISTINCT cs.linker_idlinker
FROM linkerSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.linker_idlinker IN (sqlc.slice('ids'))
;

-- name: addToForumBlogSearch :exec
INSERT IGNORE INTO blogsSearch
(blogs_idblogs, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: blogsSearchDelete :exec
DELETE FROM blogsSearch
WHERE blogs_idblogs=?
;

-- name: blogsSearchFirst :many
SELECT DISTINCT cs.blogs_idblogs
FROM blogsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: blogsSearchNext :many
SELECT DISTINCT cs.blogs_idblogs
FROM blogsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.blogs_idblogs IN (sqlc.slice('ids'))
;

-- name: topicAllowThis :one
SELECT r.*, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic=r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic=t.idforumtopic AND u.users_idusers=?
WHERE t.idforumtopic=? LIMIT 1;

-- name: threadAllowThis :one
SELECT r.*, u.level FROM forumthread t
LEFT JOIN topicrestrictions r ON t.forumtopic_idforumtopic=r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic=t.forumtopic_idforumtopic AND u.users_idusers=?
WHERE t.idforumthread=? LIMIT 1;

-- name: makePost :execlastid
INSERT INTO comments (language_idlanguage, users_idusers, forumthread_idforumthread, text, written)
VALUES (?, ?, ?, ?, NOW());

-- name: getComment :one
SELECT c.*
FROM comments c
WHERE c.Idcomments=?;

-- name: getComments :many
SELECT c.*
FROM comments c
WHERE c.Idcomments IN (sqlc.slice('ids'))
;

-- name: getCommentsWithThreadInfo :many
SELECT c.*, pu.username AS posterusername, th.idforumthread, t.idforumtopic, t.title AS forumtopic_title, fc.idforumcategory, fc.title AS forumcategory_title
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.Idcomments IN (sqlc.slice('ids'))
ORDER BY c.written DESC;
;

-- name: makeThread :execlastid
INSERT INTO forumthread (forumtopic_idforumtopic) VALUES (?);

-- name: makeCategory :exec
INSERT INTO forumcategory (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: makeTopic :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: findForumTopicByName :one
SELECT idforumtopic FROM forumtopic WHERE title=?;

-- name: user_get_all_comments_for_thread :many
SELECT c.*, pu.username AS posterusername
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_idforumthread=? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY c.written;

-- name: user_get_threads_for_topic :many
SELECT th.*, lu.username AS lastposterusername, lu.idusers AS lastposterid, fcu.username as firstpostusername, fc.written as firstpostwritten, fc.text as firstposttext
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
LEFT JOIN comments fc ON th.firstpost=fc.idcomments
LEFT JOIN users fcu ON fcu.idusers = fc.users_idusers
WHERE th.forumtopic_idforumtopic=? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY th.lastaddition DESC;

-- name: somethingNotifyBlogs :many
SELECT u.email FROM blogs t, users u, preferences p
WHERE t.idblogs=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: somethingNotifyLinker :many
SELECT u.email FROM linker t, users u, preferences p
WHERE t.idlinker=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: somethingNotifyWriting :many
SELECT u.email FROM writing t, users u, preferences p
WHERE t.idwriting=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: threadNotify :many
SELECT u.email FROM comments c, users u, preferences p
WHERE c.forumthread_idforumthread=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: deleteUserLanguage :exec
DELETE FROM userlang WHERE users_idusers = ?;

-- name: fetchLanguages :many
SELECT idlanguage, nameof FROM language;

-- -- name: updateOrInsertUserLanguage :exec
-- WITH pref_count AS (
--   SELECT COUNT(users_idusers) AS prefcount FROM preferences WHERE users_idusers = ?
-- )
-- INSERT INTO preferences (language_idlanguage, users_idusers)
-- VALUES (?, ?)
-- ON DUPLICATE KEY UPDATE
--   language_idlanguage = VALUES(language_idlanguage);

-- name: fetchUserLanguagePreferences :many
SELECT idlanguage, nameof, (
  SELECT COUNT(sul.iduserlang) FROM userlang sul
  WHERE sul.language_idlanguage = l.idlanguage AND sul.users_idusers = ?
) AS user_lang_pref
FROM language l;

-- -- name: updateOrInsertEmailForumUpdates :exec
-- WITH email_updates AS (
--   SELECT emailforumupdates FROM preferences WHERE users_idusers = ?
-- )
-- INSERT INTO preferences (emailforumupdates, users_idusers)
-- VALUES (?, ?)
-- ON DUPLICATE KEY UPDATE
--   emailforumupdates = VALUES(emailforumupdates);

-- name: fetchUserEmailForumUpdates :many
SELECT emailforumupdates FROM preferences WHERE users_idusers = ?;

-- name: assignWritingThisThreadId :exec
UPDATE writing SET forumthread_idforumthread = ? WHERE idwriting = ?;

-- name: fetchPublicWritings :many
SELECT w.title, w.abstract, w.idwriting, w.private, w.writingCategory_idwritingCategory
FROM writing w
WHERE w.private = 0
ORDER BY w.published DESC LIMIT 15;

-- name: fetchPublicWritingsInCategory :many
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_idforumthread=w.forumthread_idforumthread AND w.forumthread_idforumthread != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writingCategory_idwritingCategory=?
ORDER BY w.published DESC LIMIT 15;

-- name: updateWriting :exec
UPDATE writing
SET title = ?, abstract = ?, writting = ?, private = ?, language_idlanguage = ?
WHERE idwriting = ?;

-- name: insertWriting :execlastid
INSERT INTO writing (writingCategory_idwritingCategory, title, abstract, writting, private, language_idlanguage, published, users_idusers)
VALUES (?, ?, ?, ?, ?, ?, NOW(), ?);

-- name: fetchWritingById :one
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writtingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = sqlc.arg(UserId)
WHERE w.idwriting = ? AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = sqlc.arg(UserId))
ORDER BY w.published DESC
;

-- name: fetchWritingByIds :many
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writtingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = sqlc.arg(userId)
WHERE w.idwriting IN (sqlc.slice(writingIds)) AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = sqlc.arg(userId))
ORDER BY w.published DESC
;

-- name: fetchPublicWritingsByCategory :many
SELECT w.title, w.abstract, u.username, w.published, w.idwriting, w.private, IF(th.comments IS NULL, 0, th.comments + 1)
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN forumthread th ON w.forumthread_idforumthread = th.idforumthread
LEFT JOIN writtingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = ?
WHERE w.writingCategory_idwritingCategory = ? AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = ?)
ORDER BY w.published DESC;

-- name: fetchWritingApproval :many
SELECT editdoc
FROM writtingApprovedUsers
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: fetchWritingOwner :many
SELECT users_idusers
FROM writing
WHERE idwriting = ?;

-- name: fetchWritingByIdWithEdit :many
SELECT w.title, w.abstract, w.writting, u.username, w.published, w.idwriting, w.private, wau.editdoc, w.forumthread_idforumthread,
u.idusers, w.writingCategory_idwritingCategory
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writtingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = ?
WHERE w.idwriting = ? AND w.users_idusers = ? AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = ?)
AND (wau.editdoc = 1 OR w.users_idusers = ?)
ORDER BY w.published DESC;

-- name: fetchChildCategories :many
SELECT c3.idwritingCategory, c3.title, c2.idwritingCategory, c2.title
FROM writingCategory c1
LEFT JOIN writingCategory c2 ON c2.idwritingCategory = c1.writingCategory_idwritingCategory
LEFT JOIN writingCategory c3 ON c3.idwritingCategory = c2.writingCategory_idwritingCategory
WHERE c1.idwritingCategory = ?;

-- name: insertWritingCategory :exec
INSERT INTO writingCategory (writingCategory_idwritingCategory, title, description)
VALUES (?, ?, ?);

-- name: updateWritingCategory :exec
UPDATE writingCategory
SET title = ?, description = ?, writingCategory_idwritingCategory = ?
WHERE idwritingCategory = ?;

-- name: fetchCategories :many
SELECT idwritingCategory, title, description
FROM writingCategory
WHERE writingCategory_idwritingCategory = ?;

-- name: fetchAllCategories :many
SELECT wc.*
FROM writingCategory wc
;

-- name: deleteWritingApproval :exec
DELETE FROM writtingApprovedUsers
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: insertWritingApproval :exec
INSERT INTO writtingApprovedUsers (writing_idwriting, users_idusers, readdoc, editdoc)
VALUES (?, ?, ?, ?);

-- name: updateWritingApproval :exec
UPDATE writtingApprovedUsers
SET readdoc = ?, editdoc = ?
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: fetchWritingApprovals :many
SELECT idusers, u.username, wau.readdoc, wau.editdoc
FROM writtingApprovedUsers wau
LEFT JOIN users u ON idusers = wau.users_idusers
WHERE writing_idwriting = ?;

-- name: fetchAllWritingApprovals :many
SELECT idusers, u.username, wau.writing_idwriting, wau.readdoc, wau.editdoc
FROM writtingApprovedUsers wau
LEFT JOIN users u ON idusers = wau.users_idusers
;

-- name: fetchPagePermissions :many
SELECT p.idpermissions, p.level, u.username, u.email, p.section
FROM permissions p
JOIN users u ON u.idusers = p.users_idusers
WHERE p.section = ?
ORDER BY p.level;

-- name: insertPagePermission :exec
INSERT INTO permissions (users_idusers, section, level)
VALUES (?, ?, ?);

-- name: deletePagePermission :exec
DELETE FROM permissions WHERE idpermissions = ? AND section = ?;

-- name: updateWritingForumThreadId :exec
UPDATE writing
SET forumthread_idforumthread = ?
WHERE idwriting = ?;

-- name: Login :one
SELECT *
FROM users
WHERE username = ? AND passwd = md5(?);

-- name: userByUid :one
SELECT *
FROM users
WHERE idusers = ?;

-- name: UserByUsername :one
SELECT *
FROM users
WHERE username = ?;

-- name: UserByEmail :one
SELECT *
FROM users
WHERE email = ?;

-- name: CheckExistingUser :one
SELECT username FROM users WHERE username = ?;


-- name: InsertUser :execresult
INSERT INTO users (username, passwd, email)
VALUES (?, MD5(?), ?)
;

-- name: blogsUserPermissions :many
SELECT p.idpermissions, p.level, u.username, u.email, p.section
FROM permissions p, users u
WHERE u.idusers = p.users_idusers AND p.section = "blogs"
ORDER BY p.level
;

-- name: SelectUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

-- name: AllQuestions :many
SELECT *
FROM faq;
