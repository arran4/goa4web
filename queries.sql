-- name: GetUsersTopicLevel :one
SELECT utl.*
FROM userstopiclevel utl
WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?
;

-- name: AllUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.*
FROM users u;

-- name: CompleteWordList :many
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: RemakeCommentsSearch :exec
-- This query selects data from the "comments" table and populates the "commentsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "commentsSearch" using the "comments_idcomments".
INSERT INTO commentsSearch (text, comments_idcomments)
SELECT text, idcomments
FROM comments;

-- name: RemakeNewsSearch :exec
-- This query selects data from the "siteNews" table and populates the "siteNewsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "siteNews_idsiteNews".
INSERT INTO siteNewsSearch (text, siteNews_idsiteNews)
SELECT news, idsiteNews
FROM siteNews;

-- name: RemakeBlogSearch :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blogs_idblogs".
INSERT INTO blogsSearch (text, blogs_idblogs)
SELECT blog, idblogs
FROM blogs;

-- name: RemakeWritingSearch :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_idwriting".
INSERT INTO writingSearch (text, writing_idwriting)
SELECT CONCAT(title, ' ', abstract, ' ', writting), idwriting
FROM writing;

-- name: RemakeLinkerSearch :exec
-- This query selects data from the "linker" table and populates the "linkerSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linkerSearch" using the "linker_idlinker".
INSERT INTO linkerSearch (text, linker_idlinker)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;

-- name: RemakeCommentsSearchInsert :exec
-- This query selects data from the "comments" table and populates the "commentsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "commentsSearch" using the "comments_idcomments".
INSERT INTO commentsSearch (text, comments_idcomments)
SELECT text, idcomments
FROM comments;

-- name: DeleteCommentsSearch :exec
-- This query deletes all data from the "commentsSearch" table.
DELETE FROM commentsSearch;

-- name: RemakeNewsSearchInsert :exec
-- This query selects data from the "siteNews" table and populates the "siteNewsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "siteNews_idsiteNews".
INSERT INTO siteNewsSearch (text, siteNews_idsiteNews)
SELECT news, idsiteNews
FROM siteNews;

-- name: DeleteSiteNewsSearch :exec
-- This query deletes all data from the "siteNewsSearch" table.
DELETE FROM siteNewsSearch;

-- name: RemakeBlogsSearchInsert :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blogs_idblogs".
INSERT INTO blogsSearch (text, blogs_idblogs)
SELECT blog, idblogs
FROM blogs;

-- name: DeleteBlogsSearch :exec
-- This query deletes all data from the "blogsSearch" table.
DELETE FROM blogsSearch;

-- name: RemakeWritingSearchInsert :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_idwriting".
INSERT INTO writingSearch (text, writing_idwriting)
SELECT CONCAT(title, ' ', abstract, ' ', writting), idwriting
FROM writing;

-- name: DeleteWritingSearch :exec
-- This query deletes all data from the "writingSearch" table.
DELETE FROM writingSearch;

-- name: RemakeLinkerSearchInsert :exec
-- This query selects data from the "linker" table and populates the "linkerSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linkerSearch" using the "linker_idlinker".
INSERT INTO linkerSearch (text, linker_idlinker)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;

-- name: DeleteLinkerSearch :exec
-- This query deletes all data from the "linkerSearch" table.
DELETE FROM linkerSearch;

-- name: Update_forumthreads :exec
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

-- name: Update_forumtopics :exec
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

-- name: Update_forumthread :exec
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

-- name: Update_forumtopic :exec
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

-- name: Update_blog :exec
UPDATE blogs
SET language_idlanguage = ?, blog = ?
WHERE idblogs = ?;

-- name: Update_comment :exec
UPDATE comments
SET language_idlanguage = ?, text = ?
WHERE idcomments = ?;

-- name: Add_blog :execlastid
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
VALUES (?, ?, ?, NOW());

-- name: Assign_blog_to_thread :exec
UPDATE blogs
SET forumthread_idforumthread = ?
WHERE idblogs = ?;

-- name: Show_latest_blogs :many
SELECT b.blog, b.written, u.username, b.idblogs, coalesce(th.comments, 0), b.users_idusers
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogs :many
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

-- name: Show_blog :one
SELECT b.blog, b.written, u.username, b.idblogs, coalesce(th.comments, 0), b.users_idusers, b.forumthread_idforumthread
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.idblogs = ?
LIMIT 1;

-- name: User_get_comment :one
SELECT c.*, pu.Username
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = ? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
LIMIT 1;

-- name: Show_blogger_list :many
SELECT u.username, COUNT(b.idblogs)
FROM blogs b, users u
WHERE b.users_idusers = u.idusers
GROUP BY u.idusers;

-- name: Blog_atom :many
SELECT b.idblogs, LEFT(b.written, 255), b.blog, u.username
FROM blogs b, users u
WHERE u.idusers = b.users_idusers AND b.users_idusers = ?
ORDER BY b.written DESC
LIMIT ?;

-- name: Blog_rss :many
SELECT b.idblogs, LEFT(b.written, 255), b.blog, u.username
FROM blogs b, users u
WHERE u.idusers = b.users_idusers AND b.users_idusers= ?
ORDER BY b.written DESC
LIMIT ?;

-- name: Show_blog_edit :one
SELECT b.blog, b.language_idlanguage
FROM blogs b, users u
WHERE b.users_idusers = u.idusers AND b.idblogs = ?
LIMIT 1;

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

-- name: Rename_category :exec
UPDATE faqCategories
SET name = ?
WHERE idfaqCategories = ?;

-- name: Delete_category :exec
DELETE FROM faqCategories
WHERE idfaqCategories = ?;

-- name: Create_category :exec
INSERT INTO faqCategories (name)
VALUES (?);

-- name: Add_question :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
VALUES (?, ?, ?);

-- name: Modify_faq :exec
UPDATE faq
SET answer = ?, question = ?, faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: Delete_faq :exec
DELETE FROM faq
WHERE idfaq = ?;

-- name: Faq_categories :many
SELECT idfaqCategories, name
FROM faqCategories;

-- name: Show_questions :many
SELECT c.*, f.*
FROM faq f
LEFT JOIN faqCategories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0 AND f.answer IS NOT NULL
ORDER BY c.idfaqCategories;

-- name: ChangeCategory :exec
UPDATE forumcategory SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumcategory = ?;

-- name: ShowAllCategories :many
SELECT c.*, COUNT(c2.idforumcategory) as SubcategoryCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.forumcategory_idforumcategory = c2.idforumcategory
GROUP BY c.idforumcategory;

-- name: GetAllTopics :many
SELECT t.*
FROM forumtopic t
LEFT JOIN forumcategory c ON t.forumcategory_idforumcategory = c.idforumcategory
GROUP BY t.idforumtopic;

-- name: ChangeTopic :exec
UPDATE forumtopic SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumtopic = ?;

-- name: Get_all_user_topics_for_category :many
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = ? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: Get_all_user_topics :many
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: User_get_topic :one
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0) AND t.idforumtopic=?
ORDER BY t.lastaddition DESC;

-- name: User_get_thread :one
SELECT th.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0) AND th.idforumthread=?
ORDER BY t.lastaddition DESC;

-- name: DeleteTopicRestrictions :exec
DELETE FROM topicrestrictions WHERE forumtopic_idforumtopic = ?;

-- name: SetTopicRestrictions :exec
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

-- name: GetTopicRestrictions :many
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: GetAllTopicRestrictions :many
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic;

-- name: DeleteUsersTopicLevel :exec
DELETE FROM userstopiclevel WHERE forumtopic_idforumtopic = ? AND users_idusers = ?;

-- name: SetUsersTopicLevel :exec
INSERT INTO userstopiclevel (forumtopic_idforumtopic, users_idusers, level, invitemax)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE level = VALUES(level), invitemax = VALUES(invitemax);

-- name: GetUsersAllTopicLevels :many
SELECT u.*, t.*, utl.*, tr.*
FROM users u
JOIN userstopiclevel utl ON utl.users_idusers=u.idusers
JOIN forumtopic t ON utl.forumtopic_idforumtopic = t.idforumtopic
JOIN topicrestrictions tr ON t.idforumtopic = tr.forumtopic_idforumtopic
WHERE u.idusers = ?;

-- name: GetAllUsersAllTopicLevels :many
SELECT u.*, t.*, utl.*, tr.*
FROM users u
JOIN userstopiclevel utl ON utl.users_idusers=u.idusers
JOIN forumtopic t ON utl.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN topicrestrictions tr ON t.idforumtopic = tr.forumtopic_idforumtopic;

-- name: ForumCategories :many
SELECT f.*
FROM forumcategory f;

-- name: ImageboardRSS :exec
SELECT title, description FROM imageboard WHERE idimageboard = ?;

-- name: MakeImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description) VALUES (?, ?, ?);

-- name: ChangeImageBoard :exec
UPDATE imageboard SET title = ?, description = ?, imageboard_idimageboard = ? WHERE idimageboard = ?;

-- name: PrintSubBoards :many
SELECT idimageboard, title, description FROM imageboard WHERE imageboard_idimageboard = ?;

-- name: PrintImagePosts :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.imageboard_idimageboard = ?;

-- name: PrintImagePost :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.idimagepost = ?;

-- name: AddImage :exec
INSERT INTO imagepost (imageboard_idimageboard, thumbnail, fullimage, users_idusers, description, posted)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: AssignImagePostThisThreadId :exec
UPDATE imagepost SET forumthread_idforumthread = ? WHERE idimagepost = ?;

-- name: ShowAllBoards :many
SELECT b.idimageboard, b.title, b.description, b.imageboard_idimageboard, pb.title
FROM imageboard b
LEFT JOIN imageboard pb ON b.imageboard_idimageboard = pb.idimageboard OR b.imageboard_idimageboard = 0
GROUP BY b.idimageboard;

-- name: WriteSiteNewsRSS :many
SELECT s.idsiteNews, s.occured, s.news
FROM siteNews s
ORDER BY s.occured DESC LIMIT 15;

-- name: WriteNewsPost :exec
INSERT INTO siteNews (news, users_idusers, occured, language_idlanguage)
VALUES (?, ?, NOW(), ?);

-- name: EditNewsPost :exec
UPDATE siteNews SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?;

-- name: GetNewsThreadId :one
SELECT s.forumthread_idforumthread, u.idusers
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: AssignNewsThisThreadId :exec
UPDATE siteNews SET forumthread_idforumthread = ? WHERE idsiteNews = ?;

-- name: GetNewsPost :one
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.idsiteNews = ?
;

-- name: GetNewsPosts :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds))
;

-- name: GetLatestNewsPosts :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
ORDER BY s.occured DESC
LIMIT 15
;

-- name: DeleteCategory :exec
DELETE FROM linkerCategory WHERE idlinkerCategory = ?;

-- name: RenameCategory :exec
UPDATE linkerCategory SET title = ? WHERE idlinkerCategory = ?;

-- name: CreateCategory :exec
INSERT INTO linkerCategory (title) VALUES (?);

-- name: ShowCategories :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: AdminCategories :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: DeleteQueueItem :exec
DELETE FROM linkerQueue WHERE idlinkerQueue = ?;

-- name: UpdateQueue :exec
UPDATE linkerQueue SET linkerCategory_idlinkerCategory = ?, title = ?, url = ?, description = ? WHERE idlinkerQueue = ?;

-- name: AddToQueue :exec
INSERT INTO linkerQueue (users_idusers, linkerCategory_idlinkerCategory, title, url, description) VALUES (?, ?, ?, ?, ?);

-- name: ShowAdminQueue :many
SELECT l.*, u.username, c.title as category_title, c.idlinkerCategory
FROM linkerQueue l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory c ON l.linkerCategory_idlinkerCategory = c.idlinkerCategory
;
-- name: MoveToLinker :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, language_idlanguage, title, `url`, description)
SELECT l.users_idusers, l.linkerCategory_idlinkerCategory, l.language_idlanguage, l.title, l.url, l.description
FROM linkerQueue l
WHERE l.idlinkerQueue = ?
;

-- name: AddToLinker :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, title, url, description, listed)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: AssignLinkerThisThreadId :exec
UPDATE linker SET forumthread_idforumthread = ? WHERE idlinker = ?;

-- name: ShowLatest :many
SELECT l.*, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_idforumthread = th.idforumthread
WHERE lc.idlinkerCategory = ?
ORDER BY l.listed DESC;

-- name: ShowLink :one
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker = ?;

-- name: ShowLinks :many
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds));

-- name: Usernametouid :one
SELECT idusers FROM users WHERE username = ?;

-- name: GetWordID :one
SELECT idsearchwordlist FROM searchwordlist WHERE word = lcase(?);

-- name: AddWord :execlastid
INSERT IGNORE INTO searchwordlist (word)
VALUES (lcase(sqlc.arg(word)));

-- name: AddToForumCommentSearch :exec
INSERT IGNORE INTO commentsSearch
(comments_idcomments, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: CommentsSearchFirstNotInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND ft.forumcategory_idforumcategory!=0
;

-- name: CommentsSearchNextNotInRestrictedTopic :many
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

-- name: CommentsSearchFirstInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
WHERE swl.word=?
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: CommentsSearchNextInRestrictedTopic :many
SELECT DISTINCT cs.comments_idcomments
FROM commentsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comments_idcomments
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
WHERE swl.word=?
AND cs.comments_idcomments IN (sqlc.slice('ids'))
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: AddToForumWritingSearch :exec
INSERT IGNORE INTO writingSearch
(writing_idwriting, searchwordlist_idsearchwordlist)
VALUES (?, ?);


-- name: WritingSearchDelete :exec
DELETE FROM writingSearch
WHERE writing_idwriting=?
;

-- name: WritingSearchFirst :many
SELECT DISTINCT cs.writing_idwriting
FROM writingSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: WritingSearchNext :many
SELECT DISTINCT cs.writing_idwriting
FROM writingSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.writing_idwriting IN (sqlc.slice('ids'))
;

-- name: SiteNewsSearchFirst :many
SELECT DISTINCT cs.siteNews_idsiteNews
FROM siteNewsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: SiteNewsSearchNext :many
SELECT DISTINCT cs.siteNews_idsiteNews
FROM siteNewsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.siteNews_idsiteNews IN (sqlc.slice('ids'))
;

-- name: LinkerSearchFirst :many
SELECT DISTINCT cs.linker_idlinker
FROM linkerSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: LinkerSearchNext :many
SELECT DISTINCT cs.linker_idlinker
FROM linkerSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.linker_idlinker IN (sqlc.slice('ids'))
;

-- name: BlogsSearchFirst :many
SELECT DISTINCT cs.blogs_idblogs
FROM blogsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: BlogsSearchNext :many
SELECT DISTINCT cs.blogs_idblogs
FROM blogsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.blogs_idblogs IN (sqlc.slice('ids'))
;

-- name: MakePost :execlastid
INSERT INTO comments (language_idlanguage, users_idusers, forumthread_idforumthread, text, written)
VALUES (?, ?, ?, ?, NOW());

-- name: GetComment :one
SELECT c.*
FROM comments c
WHERE c.Idcomments=?;

-- name: GetComments :many
SELECT c.*
FROM comments c
WHERE c.Idcomments IN (sqlc.slice('ids'))
;

-- name: GetCommentsWithThreadInfo :many
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

-- name: MakeThread :execlastid
INSERT INTO forumthread (forumtopic_idforumtopic) VALUES (?);

-- name: MakeCategory :exec
INSERT INTO forumcategory (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: MakeTopic :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: FindForumTopicByName :one
SELECT idforumtopic FROM forumtopic WHERE title=?;

-- name: User_get_all_comments_for_thread :many
SELECT c.*, pu.username AS posterusername
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_idforumthread=? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY c.written;

-- name: User_get_threads_for_topic :many
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

-- name: SomethingNotifyBlogs :many
SELECT u.email FROM blogs t, users u, preferences p
WHERE t.idblogs=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: SomethingNotifyLinker :many
SELECT u.email FROM linker t, users u, preferences p
WHERE t.idlinker=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: SomethingNotifyWriting :many
SELECT u.email FROM writing t, users u, preferences p
WHERE t.idwriting=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ThreadNotify :many
SELECT u.email FROM comments c, users u, preferences p
WHERE c.forumthread_idforumthread=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: AssignWritingThisThreadId :exec
UPDATE writing SET forumthread_idforumthread = ? WHERE idwriting = ?;

-- name: FetchPublicWritings :many
SELECT w.title, w.abstract, w.idwriting, w.private, w.writingCategory_idwritingCategory
FROM writing w
WHERE w.private = 0
ORDER BY w.published DESC LIMIT 15;

-- name: FetchPublicWritingsInCategory :many
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_idforumthread=w.forumthread_idforumthread AND w.forumthread_idforumthread != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writingCategory_idwritingCategory=?
ORDER BY w.published DESC LIMIT 15;

-- name: UpdateWriting :exec
UPDATE writing
SET title = ?, abstract = ?, writting = ?, private = ?, language_idlanguage = ?
WHERE idwriting = ?;

-- name: InsertWriting :execlastid
INSERT INTO writing (writingCategory_idwritingCategory, title, abstract, writting, private, language_idlanguage, published, users_idusers)
VALUES (?, ?, ?, ?, ?, ?, NOW(), ?);

-- name: FetchWritingById :one
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writtingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = sqlc.arg(UserId)
WHERE w.idwriting = ? AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = sqlc.arg(UserId))
ORDER BY w.published DESC
;

-- name: FetchWritingByIds :many
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writtingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = sqlc.arg(userId)
WHERE w.idwriting IN (sqlc.slice(writingIds)) AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = sqlc.arg(userId))
ORDER BY w.published DESC
;

-- name: FetchWritingApproval :many
SELECT editdoc
FROM writtingApprovedUsers
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: InsertWritingCategory :exec
INSERT INTO writingCategory (writingCategory_idwritingCategory, title, description)
VALUES (?, ?, ?);

-- name: UpdateWritingCategory :exec
UPDATE writingCategory
SET title = ?, description = ?, writingCategory_idwritingCategory = ?
WHERE idwritingCategory = ?;

-- name: FetchCategories :many
SELECT idwritingCategory, title, description
FROM writingCategory
WHERE writingCategory_idwritingCategory = ?;

-- name: FetchAllCategories :many
SELECT wc.*
FROM writingCategory wc
;

-- name: DeleteWritingApproval :exec
DELETE FROM writtingApprovedUsers
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: InsertWritingApproval :exec
INSERT INTO writtingApprovedUsers (writing_idwriting, users_idusers, readdoc, editdoc)
VALUES (?, ?, ?, ?);

-- name: UpdateWritingApproval :exec
UPDATE writtingApprovedUsers
SET readdoc = ?, editdoc = ?
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: FetchAllWritingApprovals :many
SELECT idusers, u.username, wau.writing_idwriting, wau.readdoc, wau.editdoc
FROM writtingApprovedUsers wau
LEFT JOIN users u ON idusers = wau.users_idusers
;

-- name: Login :one
SELECT *
FROM users
WHERE username = ? AND passwd = md5(?);

-- name: UserByUid :one
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

-- name: InsertUser :execresult
INSERT INTO users (username, passwd, email)
VALUES (?, MD5(?), ?)
;

-- name: SelectUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

-- name: AllQuestions :many
SELECT *
FROM faq;
