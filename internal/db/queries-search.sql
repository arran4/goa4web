-- name: AdminCompleteWordList :many
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

-- name: AdminWordListWithCounts :many
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

-- name: AdminCountWordList :one
SELECT COUNT(*)
FROM searchwordlist;

-- name: AdminCountWordListByPrefix :one
SELECT COUNT(*)
FROM searchwordlist
WHERE word LIKE CONCAT(sqlc.arg(prefix), '%');

-- name: AdminWordListWithCountsByPrefix :many
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







-- name: SystemDeleteCommentsSearch :exec
-- This query deletes all data from the "comments_search" table.
DELETE FROM comments_search;


-- name: SystemDeleteSiteNewsSearch :exec
-- This query deletes all data from the "site_news_search" table.
DELETE FROM site_news_search;


-- name: SystemDeleteBlogsSearch :exec
-- This query deletes all data from the "blogs_search" table.
DELETE FROM blogs_search;


-- name: SystemDeleteWritingSearch :exec
-- This query deletes all data from the "writing_search" table.
DELETE FROM writing_search;


-- name: SystemDeleteLinkerSearch :exec
-- This query deletes all data from the "linker_search" table.
DELETE FROM linker_search;

-- name: SystemGetSearchWordByWordLowercased :one
SELECT *
FROM searchwordlist
WHERE word = lcase(?);

-- name: SystemCreateSearchWord :execlastid
INSERT INTO searchwordlist (word)
VALUES (lcase(sqlc.arg(word)))
ON DUPLICATE KEY UPDATE idsearchwordlist=LAST_INSERT_ID(idsearchwordlist);

-- name: SystemAddToForumCommentSearch :exec
INSERT INTO comments_search
(comment_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);


-- name: ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=sqlc.arg(word)
  AND ft.forumcategory_idforumcategory!=0
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = c.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE (g.section='forum' OR g.section='privateforum')
        AND (g.item='topic' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = ft.idforumtopic OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListCommentIDsBySearchWordNextForListerNotInRestrictedTopic :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=sqlc.arg(word)
  AND cs.comment_id IN (sqlc.slice('ids'))
  AND ft.forumcategory_idforumcategory!=0
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = c.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE (g.section='forum' OR g.section='privateforum')
        AND (g.item='topic' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = ft.idforumtopic OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListCommentIDsBySearchWordFirstForListerInRestrictedTopic :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=sqlc.arg(word)
  AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = c.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE (g.section='forum' OR g.section='privateforum')
        AND (g.item='topic' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = ft.idforumtopic OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListCommentIDsBySearchWordNextForListerInRestrictedTopic :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=sqlc.arg(word)
  AND cs.comment_id IN (sqlc.slice('ids'))
  AND fth.forumtopic_idforumtopic IN (sqlc.slice('ftids'))
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = c.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE (g.section='forum' OR g.section='privateforum')
        AND (g.item='topic' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = ft.idforumtopic OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );
-- name: SystemAddToForumWritingSearch :exec
INSERT INTO writing_search
(writing_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: SystemAddToLinkerSearch :exec
INSERT INTO linker_search
(linker_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);
-- name: SystemAddToBlogsSearch :exec
INSERT INTO blogs_search
(blog_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: SystemAddToSiteNewsSearch :exec
INSERT INTO site_news_search
(site_news_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);



-- name: SystemDeleteWritingSearchByWritingID :exec
DELETE FROM writing_search
WHERE writing_id = sqlc.arg(writing_id);
-- name: ListWritingSearchFirstForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.writing_id
FROM writing_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN writing w ON w.idwriting = cs.writing_id
WHERE swl.word = sqlc.arg(word)
  AND (
      w.language_id = 0
      OR w.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = w.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='writing'
        AND (g.item='article' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = w.idwriting OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListWritingSearchNextForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.writing_id
FROM writing_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN writing w ON w.idwriting = cs.writing_id
WHERE swl.word = sqlc.arg(word)
  AND cs.writing_id IN (sqlc.slice('ids'))
  AND (
      w.language_id = 0
      OR w.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = w.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='writing'
        AND (g.item='article' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = w.idwriting OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListSiteNewsSearchFirstForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.site_news_id
FROM site_news_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN site_news sn ON sn.idsiteNews = cs.site_news_id
WHERE swl.word = sqlc.arg(word)
  AND (
      sn.language_id = 0
      OR sn.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = sn.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='news'
        AND (g.item='post' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = sn.idsiteNews OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListSiteNewsSearchNextForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.site_news_id
FROM site_news_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN site_news sn ON sn.idsiteNews = cs.site_news_id
WHERE swl.word = sqlc.arg(word)
  AND cs.site_news_id IN (sqlc.slice('ids'))
  AND (
      sn.language_id = 0
      OR sn.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = sn.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='news'
        AND (g.item='post' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = sn.idsiteNews OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );



-- name: LinkerSearchFirst :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.linker_id
FROM linker_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN linker l ON l.id = cs.linker_id
WHERE swl.word = sqlc.arg(word)
  AND (
      l.language_id = 0
      OR l.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = l.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='linker'
        AND (g.item='link' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = l.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: LinkerSearchNext :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT DISTINCT cs.linker_id
FROM linker_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN linker l ON l.id = cs.linker_id
WHERE swl.word = sqlc.arg(word)
  AND cs.linker_id IN (sqlc.slice('ids'))
  AND (
      l.language_id = 0
      OR l.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_id = l.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='linker'
        AND (g.item='link' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = l.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );
-- name: SystemAddToImagePostSearch :exec
INSERT INTO imagepost_search
(image_post_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count);

-- name: SystemDeleteImagePostSearch :exec
DELETE FROM imagepost_search;



