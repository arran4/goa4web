-- name: CompleteWordList :many
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: WordListWithCounts :many
-- Show each search word with total usage counts across all search tables.
SELECT swl.word,
       (SELECT IFNULL(SUM(cs.word_count),0) FROM comments_search cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ns.word_count),0) FROM site_news_search ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(bs.word_count),0) FROM blogs_search bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ls.word_count),0) FROM linker_search ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ws.word_count),0) FROM writing_search ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ips.word_count),0) FROM imagepost_search ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
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
       (SELECT IFNULL(SUM(cs.word_count),0) FROM comments_search cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ns.word_count),0) FROM site_news_search ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(bs.word_count),0) FROM blogs_search bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ls.word_count),0) FROM linker_search ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ws.word_count),0) FROM writing_search ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ips.word_count),0) FROM imagepost_search ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
FROM searchwordlist swl
WHERE swl.word LIKE CONCAT(sqlc.arg(prefix), '%')
ORDER BY swl.word
LIMIT ? OFFSET ?;







-- name: DeleteCommentsSearch :exec
-- This query deletes all data from the "comments_search" table.
DELETE FROM comments_search;


-- name: DeleteSiteNewsSearch :exec
-- This query deletes all data from the "site_news_search" table.
DELETE FROM site_news_search;


-- name: DeleteBlogsSearch :exec
-- This query deletes all data from the "blogs_search" table.
DELETE FROM blogs_search;


-- name: DeleteWritingSearch :exec
-- This query deletes all data from the "writing_search" table.
DELETE FROM writing_search;


-- name: DeleteLinkerSearch :exec
-- This query deletes all data from the "linker_search" table.
DELETE FROM linker_search;

-- name: GetSearchWordByWordLowercased :one
SELECT *
FROM searchwordlist
WHERE word = lcase(?);

-- name: CreateSearchWord :execlastid
INSERT INTO searchwordlist (word)
VALUES (lcase(sqlc.arg(word)))
ON DUPLICATE KEY UPDATE idsearchwordlist=LAST_INSERT_ID(idsearchwordlist);

-- name: AddToForumCommentSearch :exec
INSERT INTO comments_search
(comment_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: CommentsSearchFirstNotInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND ft.forumcategory_idforumcategory!=0
;

-- name: CommentsSearchNextNotInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
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
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
WHERE swl.word=?
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: CommentsSearchNextInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
WHERE swl.word=?
AND cs.comment_id IN (sqlc.slice('ids'))
AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
;

-- name: AddToForumWritingSearch :exec
INSERT INTO writing_search
(writing_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: AddToLinkerSearch :exec
INSERT INTO linker_search
(linker_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);
-- name: AddToBlogsSearch :exec
INSERT INTO blogs_search
(blog_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: AddToSiteNewsSearch :exec
INSERT INTO site_news_search
(site_news_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);



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
INSERT INTO imagepost_search
(image_post_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: DeleteImagePostSearch :exec
DELETE FROM imagepost_search;


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

