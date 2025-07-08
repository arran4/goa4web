-- name: CompleteWordList :many
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: WordListWithCounts :many
-- Show each search word with total usage counts across all search tables.
SELECT swl.word,
       (SELECT COUNT(*) FROM commentsSearch cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM siteNewsSearch ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM blogsSearch bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM linkerSearch ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM writingSearch ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM imagepostSearch ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
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
       (SELECT COUNT(*) FROM commentsSearch cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM siteNewsSearch ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM blogsSearch bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM linkerSearch ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM writingSearch ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT COUNT(*) FROM imagepostSearch ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
FROM searchwordlist swl
WHERE swl.word LIKE CONCAT(sqlc.arg(prefix), '%')
ORDER BY swl.word
LIMIT ? OFFSET ?;

-- name: RemakeCommentsSearch :exec
-- This query selects data from the "comments" table and populates the "commentsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "commentsSearch" using the "comments_idcomments".
INSERT INTO commentsSearch (text, comments_idcomments)
SELECT text, idcomments
FROM comments;

-- name: RemakeNewsSearch :exec
-- This query selects data from the "siteNews" table and populates the "siteNewsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "site_news_id".
INSERT INTO siteNewsSearch (text, site_news_id)
SELECT news, idsiteNews
FROM siteNews;

-- name: RemakeBlogSearch :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blog_id".
INSERT INTO blogsSearch (text, blog_id)
SELECT blog, idblogs
FROM blogs;

-- name: RemakeWritingSearch :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_id".
INSERT INTO writingSearch (text, writing_id)
SELECT CONCAT(title, ' ', abstract, ' ', writing), idwriting
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
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "siteNewsSearch" using the "site_news_id".
INSERT INTO siteNewsSearch (text, site_news_id)
SELECT news, idsiteNews
FROM siteNews;

-- name: DeleteSiteNewsSearch :exec
-- This query deletes all data from the "siteNewsSearch" table.
DELETE FROM siteNewsSearch;

-- name: RemakeBlogsSearchInsert :exec
-- This query selects data from the "blogs" table and populates the "blogsSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "blogsSearch" using the "blog_id".
INSERT INTO blogsSearch (text, blog_id)
SELECT blog, idblogs
FROM blogs;

-- name: DeleteBlogsSearch :exec
-- This query deletes all data from the "blogsSearch" table.
DELETE FROM blogsSearch;

-- name: RemakeWritingSearchInsert :exec
-- This query selects data from the "writing" table and populates the "writingSearch" table with the specified columns.
-- Then, it iterates over the "queue" linked list to add each text and ID pair to the "writingSearch" using the "writing_id".
INSERT INTO writingSearch (text, writing_id)
SELECT CONCAT(title, ' ', abstract, ' ', writing), idwriting
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

-- name: GetSearchWordByWordLowercased :one
SELECT *
FROM searchwordlist
WHERE word = lcase(?);

-- name: CreateSearchWord :execlastid
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
(writing_id, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: AddToLinkerSearch :exec
INSERT IGNORE INTO linkerSearch
(linker_idlinker, searchwordlist_idsearchwordlist)
VALUES (?, ?);


-- name: WritingSearchDelete :exec
DELETE FROM writingSearch
WHERE writing_id=?
;

-- name: WritingSearchFirst :many
SELECT DISTINCT cs.writing_id
FROM writingSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: WritingSearchNext :many
SELECT DISTINCT cs.writing_id
FROM writingSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.writing_id IN (sqlc.slice('ids'))
;

-- name: SiteNewsSearchFirst :many
SELECT DISTINCT cs.site_news_id
FROM siteNewsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: SiteNewsSearchNext :many
SELECT DISTINCT cs.site_news_id
FROM siteNewsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.site_news_id IN (sqlc.slice('ids'))
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

-- name: AddToImagePostSearch :exec
INSERT IGNORE INTO imagepostSearch
(imagepost_idimagepost, searchwordlist_idsearchwordlist)
VALUES (?, ?);

-- name: DeleteImagePostSearch :exec
DELETE FROM imagepostSearch;

-- name: RemakeImagePostSearchInsert :exec
INSERT INTO imagepostSearch (text, imagepost_idimagepost)
SELECT description, idimagepost
FROM imagepost;

-- name: ImagePostSearchFirst :many
SELECT DISTINCT cs.imagepost_idimagepost
FROM imagepostSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?;

-- name: ImagePostSearchNext :many
SELECT DISTINCT cs.imagepost_idimagepost
FROM imagepostSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.imagepost_idimagepost IN (sqlc.slice('ids'));

