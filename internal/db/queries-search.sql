-- name: CompleteWordList :many
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: WordListWithCounts :many
-- Show each search word with total usage counts across all search tables.
SELECT swl.word,
       (SELECT COUNT(*) FROM comments_search cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM site_news_search ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM blogs_search bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM linker_search ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM writing_search ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM imagepost_search ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
FROM searchwordlist swl
ORDER BY swl.word
LIMIT ? OFFSET ?;

-- name: CountWordList :one
SELECT COUNT(*)
FROM searchwordlist;

-- name: CountWordListByPrefix :one
SELECT COUNT(*)
FROM searchwordlist
WHERE word LIKE CONCAT(sqlc.arg(prefix), '%');

-- name: WordListWithCountsByPrefix :many
SELECT swl.word,
       (SELECT COUNT(*) FROM comments_search cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM site_news_search ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM blogs_search bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM linker_search ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM writing_search ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM imagepost_search ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
FROM searchwordlist swl
WHERE swl.word LIKE CONCAT(sqlc.arg(prefix), '%')
ORDER BY swl.word
LIMIT ? OFFSET ?;

-- name: RemakeCommentsSearch :exec
-- This query selects data from the "comments" table and populates the "comments_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "comments_search" using the "comment_id".
INSERT INTO comments_search (text, comment_id)
SELECT text, idcomments
FROM comments;

-- name: RemakeNewsSearch :exec
-- This query selects data from the "site_news" table and populates the "site_news_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "site_news_search" using the "site_news_id".
INSERT INTO site_news_search (text, site_news_id)
SELECT news, idsiteNews
FROM site_news;

-- name: RemakeBlogSearch :exec
-- This query selects data from the "blogs" table and populates the "blogs_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogs_search" using the "blog_id".
INSERT INTO blogs_search (text, blog_id)
SELECT blog, idblogs
FROM blogs;

-- name: RemakeWritingSearch :exec
-- This query selects data from the "writing" table and populates the "writing_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writing_search" using the "writing_id".
INSERT INTO writing_search (text, writing_id)
SELECT CONCAT(title, ' ', abstract, ' ', writing), idwriting
FROM writing;

-- name: RemakeLinkerSearch :exec
-- This query selects data from the "linker" table and populates the "linker_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linker_search" using the "linker_id".
INSERT INTO linker_search (text, linker_id)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;

-- name: RemakeCommentsSearchInsert :exec
-- This query selects data from the "comments" table and populates the "comments_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "comments_search" using the "comment_id".
INSERT INTO comments_search (text, comment_id)
SELECT text, idcomments
FROM comments;

-- name: DeleteCommentsSearch :exec
-- This query deletes all data from the "comments_search" table.
DELETE FROM comments_search;

-- name: RemakeNewsSearchInsert :exec
-- This query selects data from the "site_news" table and populates the "site_news_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "site_news_search" using the "site_news_id".
INSERT INTO site_news_search (text, site_news_id)
SELECT news, idsiteNews
FROM site_news;

-- name: DeleteSiteNewsSearch :exec
-- This query deletes all data from the "site_news_search" table.
DELETE FROM site_news_search;

-- name: RemakeBlogsSearchInsert :exec
-- This query selects data from the "blogs" table and populates the "blogs_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogs_search" using the "blog_id".
INSERT INTO blogs_search (text, blog_id)
SELECT blog, idblogs
FROM blogs;

-- name: DeleteBlogsSearch :exec
-- This query deletes all data from the "blogs_search" table.
DELETE FROM blogs_search;

-- name: RemakeWritingSearchInsert :exec
-- This query selects data from the "writing" table and populates the "writing_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writing_search" using the "writing_id".
INSERT INTO writing_search (text, writing_id)
SELECT CONCAT(title, ' ', abstract, ' ', writing), idwriting
FROM writing;

-- name: DeleteWritingSearch :exec
-- This query deletes all data from the "writing_search" table.
DELETE FROM writing_search;

-- name: RemakeLinkerSearchInsert :exec
-- This query selects data from the "linker" table and populates the "linker_search" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "linker_search" using the "linker_id".
INSERT INTO linker_search (text, linker_id)
SELECT CONCAT(title, ' ', description), idlinker
FROM linker;

-- name: DeleteLinkerSearch :exec
-- This query deletes all data from the "linker_search" table.
DELETE FROM linker_search;

-- name: GetSearchWordByWordLowercased :one
SELECT *
FROM searchwordlist
WHERE word = lcase(?);

-- name: CreateSearchWord :execlastid
INSERT IGNORE INTO searchwordlist (word)
VALUES (lcase(sqlc.arg(word)));

-- name: AddToForumCommentSearch :exec
INSERT IGNORE INTO comments_search
(comment_id, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: CommentsSearchFirstNotInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND ft.forumcategory_idforumcategory!=0
;

-- name: CommentsSearchNextNotInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND cs.comment_id IN (sqlc.slice('ids'))
AND ft.forumcategory_idforumcategory!=0
;

-- name: CommentsSearchFirstInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
WHERE swl.word=?
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: CommentsSearchNextInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_idforumthread
WHERE swl.word=?
AND cs.comment_id IN (sqlc.slice('ids'))
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: AddToForumWritingSearch :exec
INSERT IGNORE INTO writing_search
(writing_id, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: AddToLinkerSearch :exec
INSERT IGNORE INTO linker_search
(linker_id, searchwordlist_idsearchwordlist)
VALUES (?, ?);


-- name: WritingSearchDelete :exec
DELETE FROM writing_search
WHERE writing_id=?
;

-- name: WritingSearchFirst :many
SELECT DISTINCT cs.writing_id
FROM writing_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: WritingSearchNext :many
SELECT DISTINCT cs.writing_id
FROM writing_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.writing_id IN (sqlc.slice('ids'))
;

-- name: SiteNewsSearchFirst :many
SELECT DISTINCT cs.site_news_id
FROM site_news_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: SiteNewsSearchNext :many
SELECT DISTINCT cs.site_news_id
FROM site_news_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.site_news_id IN (sqlc.slice('ids'))
;

-- name: LinkerSearchFirst :many
SELECT DISTINCT cs.linker_id
FROM linker_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: LinkerSearchNext :many
SELECT DISTINCT cs.linker_id
FROM linker_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.linker_id IN (sqlc.slice('ids'))
;

-- name: AddToImagePostSearch :exec
INSERT IGNORE INTO imagepost_search
(image_post_id, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: DeleteImagePostSearch :exec
DELETE FROM imagepost_search;

-- name: RemakeImagePostSearchInsert :exec
INSERT INTO imagepost_search (text, image_post_id)
SELECT description, idimagepost
FROM imagepost;

-- name: ImagePostSearchFirst :many
SELECT DISTINCT cs.image_post_id
FROM imagepost_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?;

-- name: ImagePostSearchNext :many
SELECT DISTINCT cs.image_post_id
FROM imagepost_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.image_post_id IN (sqlc.slice('ids'));

