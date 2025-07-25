// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries-search.sql

package db

import (
	"context"
	"database/sql"
	"strings"
)

const addToBlogsSearch = `-- name: AddToBlogsSearch :exec
INSERT INTO blogs_search
(blog_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count)
`

type AddToBlogsSearchParams struct {
	BlogID                         int32
	SearchwordlistIdsearchwordlist int32
	WordCount                      int32
}

func (q *Queries) AddToBlogsSearch(ctx context.Context, arg AddToBlogsSearchParams) error {
	_, err := q.db.ExecContext(ctx, addToBlogsSearch, arg.BlogID, arg.SearchwordlistIdsearchwordlist, arg.WordCount)
	return err
}

const addToForumCommentSearch = `-- name: AddToForumCommentSearch :exec
INSERT INTO comments_search
(comment_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count)
`

type AddToForumCommentSearchParams struct {
	CommentID                      int32
	SearchwordlistIdsearchwordlist int32
	WordCount                      int32
}

func (q *Queries) AddToForumCommentSearch(ctx context.Context, arg AddToForumCommentSearchParams) error {
	_, err := q.db.ExecContext(ctx, addToForumCommentSearch, arg.CommentID, arg.SearchwordlistIdsearchwordlist, arg.WordCount)
	return err
}

const addToForumWritingSearch = `-- name: AddToForumWritingSearch :exec
INSERT INTO writing_search
(writing_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count)
`

type AddToForumWritingSearchParams struct {
	WritingID                      int32
	SearchwordlistIdsearchwordlist int32
	WordCount                      int32
}

func (q *Queries) AddToForumWritingSearch(ctx context.Context, arg AddToForumWritingSearchParams) error {
	_, err := q.db.ExecContext(ctx, addToForumWritingSearch, arg.WritingID, arg.SearchwordlistIdsearchwordlist, arg.WordCount)
	return err
}

const addToImagePostSearch = `-- name: AddToImagePostSearch :exec
INSERT INTO imagepost_search
(image_post_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count)
`

type AddToImagePostSearchParams struct {
	ImagePostID                    int32
	SearchwordlistIdsearchwordlist int32
	WordCount                      int32
}

func (q *Queries) AddToImagePostSearch(ctx context.Context, arg AddToImagePostSearchParams) error {
	_, err := q.db.ExecContext(ctx, addToImagePostSearch, arg.ImagePostID, arg.SearchwordlistIdsearchwordlist, arg.WordCount)
	return err
}

const addToLinkerSearch = `-- name: AddToLinkerSearch :exec
INSERT INTO linker_search
(linker_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count)
`

type AddToLinkerSearchParams struct {
	LinkerID                       int32
	SearchwordlistIdsearchwordlist int32
	WordCount                      int32
}

func (q *Queries) AddToLinkerSearch(ctx context.Context, arg AddToLinkerSearchParams) error {
	_, err := q.db.ExecContext(ctx, addToLinkerSearch, arg.LinkerID, arg.SearchwordlistIdsearchwordlist, arg.WordCount)
	return err
}

const addToSiteNewsSearch = `-- name: AddToSiteNewsSearch :exec
INSERT INTO site_news_search
(site_news_id, searchwordlist_idsearchwordlist, word_count)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE word_count=VALUES(word_count)
`

type AddToSiteNewsSearchParams struct {
	SiteNewsID                     int32
	SearchwordlistIdsearchwordlist int32
	WordCount                      int32
}

func (q *Queries) AddToSiteNewsSearch(ctx context.Context, arg AddToSiteNewsSearchParams) error {
	_, err := q.db.ExecContext(ctx, addToSiteNewsSearch, arg.SiteNewsID, arg.SearchwordlistIdsearchwordlist, arg.WordCount)
	return err
}

const commentsSearchFirstInRestrictedTopic = `-- name: CommentsSearchFirstInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
WHERE swl.word=?
AND fth.forumtopic_idforumtopic IN (/*SLICE:ftids*/?)
`

type CommentsSearchFirstInRestrictedTopicParams struct {
	Word  sql.NullString
	Ftids []int32
}

func (q *Queries) CommentsSearchFirstInRestrictedTopic(ctx context.Context, arg CommentsSearchFirstInRestrictedTopicParams) ([]int32, error) {
	query := commentsSearchFirstInRestrictedTopic
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ftids) > 0 {
		for _, v := range arg.Ftids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ftids*/?", strings.Repeat(",?", len(arg.Ftids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ftids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var comment_id int32
		if err := rows.Scan(&comment_id); err != nil {
			return nil, err
		}
		items = append(items, comment_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const commentsSearchFirstNotInRestrictedTopic = `-- name: CommentsSearchFirstNotInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND ft.forumcategory_idforumcategory!=0
`

func (q *Queries) CommentsSearchFirstNotInRestrictedTopic(ctx context.Context, word sql.NullString) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, commentsSearchFirstNotInRestrictedTopic, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var comment_id int32
		if err := rows.Scan(&comment_id); err != nil {
			return nil, err
		}
		items = append(items, comment_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const commentsSearchNextInRestrictedTopic = `-- name: CommentsSearchNextInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
WHERE swl.word=?
AND cs.comment_id IN (/*SLICE:ids*/?)
AND fth.forumtopic_idforumtopic IN (/*SLICE:ftids*/?)
`

type CommentsSearchNextInRestrictedTopicParams struct {
	Word  sql.NullString
	Ids   []int32
	Ftids []int32
}

func (q *Queries) CommentsSearchNextInRestrictedTopic(ctx context.Context, arg CommentsSearchNextInRestrictedTopicParams) ([]int32, error) {
	query := commentsSearchNextInRestrictedTopic
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	if len(arg.Ftids) > 0 {
		for _, v := range arg.Ftids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ftids*/?", strings.Repeat(",?", len(arg.Ftids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ftids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var comment_id int32
		if err := rows.Scan(&comment_id); err != nil {
			return nil, err
		}
		items = append(items, comment_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const commentsSearchNextNotInRestrictedTopic = `-- name: CommentsSearchNextNotInRestrictedTopic :many
SELECT DISTINCT cs.comment_id
FROM comments_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
LEFT JOIN comments c ON c.idcomments=cs.comment_id
LEFT JOIN forumthread fth ON fth.idforumthread=c.forumthread_id
LEFT JOIN forumtopic ft ON ft.idforumtopic=fth.forumtopic_idforumtopic
WHERE swl.word=?
AND cs.comment_id IN (/*SLICE:ids*/?)
AND ft.forumcategory_idforumcategory!=0
`

type CommentsSearchNextNotInRestrictedTopicParams struct {
	Word sql.NullString
	Ids  []int32
}

func (q *Queries) CommentsSearchNextNotInRestrictedTopic(ctx context.Context, arg CommentsSearchNextNotInRestrictedTopicParams) ([]int32, error) {
	query := commentsSearchNextNotInRestrictedTopic
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var comment_id int32
		if err := rows.Scan(&comment_id); err != nil {
			return nil, err
		}
		items = append(items, comment_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const completeWordList = `-- name: CompleteWordList :many
SELECT word
FROM searchwordlist
`

// This query selects all words from the "searchwordlist" table and prints them.
func (q *Queries) CompleteWordList(ctx context.Context) ([]sql.NullString, error) {
	rows, err := q.db.QueryContext(ctx, completeWordList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []sql.NullString
	for rows.Next() {
		var word sql.NullString
		if err := rows.Scan(&word); err != nil {
			return nil, err
		}
		items = append(items, word)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const countWordList = `-- name: CountWordList :one
SELECT COUNT(*)
FROM searchwordlist
`

func (q *Queries) CountWordList(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countWordList)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countWordListByPrefix = `-- name: CountWordListByPrefix :one
SELECT COUNT(*)
FROM searchwordlist
WHERE word LIKE CONCAT(?, '%')
`

func (q *Queries) CountWordListByPrefix(ctx context.Context, prefix interface{}) (int64, error) {
	row := q.db.QueryRowContext(ctx, countWordListByPrefix, prefix)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createSearchWord = `-- name: CreateSearchWord :execlastid
INSERT IGNORE INTO searchwordlist (word)
VALUES (lcase(?))
`

func (q *Queries) CreateSearchWord(ctx context.Context, word string) (int64, error) {
	result, err := q.db.ExecContext(ctx, createSearchWord, word)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const deleteBlogsSearch = `-- name: DeleteBlogsSearch :exec
DELETE FROM blogs_search
`

// This query deletes all data from the "blogs_search" table.
func (q *Queries) DeleteBlogsSearch(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteBlogsSearch)
	return err
}

const deleteCommentsSearch = `-- name: DeleteCommentsSearch :exec
DELETE FROM comments_search
`

// This query deletes all data from the "comments_search" table.
func (q *Queries) DeleteCommentsSearch(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteCommentsSearch)
	return err
}

const deleteImagePostSearch = `-- name: DeleteImagePostSearch :exec
DELETE FROM imagepost_search
`

func (q *Queries) DeleteImagePostSearch(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteImagePostSearch)
	return err
}

const deleteLinkerSearch = `-- name: DeleteLinkerSearch :exec
DELETE FROM linker_search
`

// This query deletes all data from the "linker_search" table.
func (q *Queries) DeleteLinkerSearch(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteLinkerSearch)
	return err
}

const deleteSiteNewsSearch = `-- name: DeleteSiteNewsSearch :exec
DELETE FROM site_news_search
`

// This query deletes all data from the "site_news_search" table.
func (q *Queries) DeleteSiteNewsSearch(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteSiteNewsSearch)
	return err
}

const deleteWritingSearch = `-- name: DeleteWritingSearch :exec
DELETE FROM writing_search
`

// This query deletes all data from the "writing_search" table.
func (q *Queries) DeleteWritingSearch(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteWritingSearch)
	return err
}

const getSearchWordByWordLowercased = `-- name: GetSearchWordByWordLowercased :one
SELECT idsearchwordlist, word
FROM searchwordlist
WHERE word = lcase(?)
`

func (q *Queries) GetSearchWordByWordLowercased(ctx context.Context, lcase string) (*Searchwordlist, error) {
	row := q.db.QueryRowContext(ctx, getSearchWordByWordLowercased, lcase)
	var i Searchwordlist
	err := row.Scan(&i.Idsearchwordlist, &i.Word)
	return &i, err
}

const imagePostSearchFirst = `-- name: ImagePostSearchFirst :many
SELECT DISTINCT cs.image_post_id
FROM imagepost_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
`

func (q *Queries) ImagePostSearchFirst(ctx context.Context, word sql.NullString) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, imagePostSearchFirst, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var image_post_id int32
		if err := rows.Scan(&image_post_id); err != nil {
			return nil, err
		}
		items = append(items, image_post_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const imagePostSearchNext = `-- name: ImagePostSearchNext :many
SELECT DISTINCT cs.image_post_id
FROM imagepost_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.image_post_id IN (/*SLICE:ids*/?)
`

type ImagePostSearchNextParams struct {
	Word sql.NullString
	Ids  []int32
}

func (q *Queries) ImagePostSearchNext(ctx context.Context, arg ImagePostSearchNextParams) ([]int32, error) {
	query := imagePostSearchNext
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var image_post_id int32
		if err := rows.Scan(&image_post_id); err != nil {
			return nil, err
		}
		items = append(items, image_post_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const linkerSearchFirst = `-- name: LinkerSearchFirst :many
SELECT DISTINCT cs.linker_id
FROM linker_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
`

func (q *Queries) LinkerSearchFirst(ctx context.Context, word sql.NullString) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, linkerSearchFirst, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var linker_id int32
		if err := rows.Scan(&linker_id); err != nil {
			return nil, err
		}
		items = append(items, linker_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const linkerSearchNext = `-- name: LinkerSearchNext :many
SELECT DISTINCT cs.linker_id
FROM linker_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.linker_id IN (/*SLICE:ids*/?)
`

type LinkerSearchNextParams struct {
	Word sql.NullString
	Ids  []int32
}

func (q *Queries) LinkerSearchNext(ctx context.Context, arg LinkerSearchNextParams) ([]int32, error) {
	query := linkerSearchNext
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var linker_id int32
		if err := rows.Scan(&linker_id); err != nil {
			return nil, err
		}
		items = append(items, linker_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const siteNewsSearchFirst = `-- name: SiteNewsSearchFirst :many
SELECT DISTINCT cs.site_news_id
FROM site_news_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
`

func (q *Queries) SiteNewsSearchFirst(ctx context.Context, word sql.NullString) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, siteNewsSearchFirst, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var site_news_id int32
		if err := rows.Scan(&site_news_id); err != nil {
			return nil, err
		}
		items = append(items, site_news_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const siteNewsSearchNext = `-- name: SiteNewsSearchNext :many
SELECT DISTINCT cs.site_news_id
FROM site_news_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.site_news_id IN (/*SLICE:ids*/?)
`

type SiteNewsSearchNextParams struct {
	Word sql.NullString
	Ids  []int32
}

func (q *Queries) SiteNewsSearchNext(ctx context.Context, arg SiteNewsSearchNextParams) ([]int32, error) {
	query := siteNewsSearchNext
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var site_news_id int32
		if err := rows.Scan(&site_news_id); err != nil {
			return nil, err
		}
		items = append(items, site_news_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const wordListWithCounts = `-- name: WordListWithCounts :many
SELECT swl.word,
       (SELECT IFNULL(SUM(cs.word_count),0) FROM comments_search cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ns.word_count),0) FROM site_news_search ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(bs.word_count),0) FROM blogs_search bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ls.word_count),0) FROM linker_search ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ws.word_count),0) FROM writing_search ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ips.word_count),0) FROM imagepost_search ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
FROM searchwordlist swl
ORDER BY swl.word
LIMIT ? OFFSET ?
`

type WordListWithCountsParams struct {
	Limit  int32
	Offset int32
}

type WordListWithCountsRow struct {
	Word  sql.NullString
	Count int32
}

// Show each search word with total usage counts across all search tables.
func (q *Queries) WordListWithCounts(ctx context.Context, arg WordListWithCountsParams) ([]*WordListWithCountsRow, error) {
	rows, err := q.db.QueryContext(ctx, wordListWithCounts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*WordListWithCountsRow
	for rows.Next() {
		var i WordListWithCountsRow
		if err := rows.Scan(&i.Word, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const wordListWithCountsByPrefix = `-- name: WordListWithCountsByPrefix :many
SELECT swl.word,
       (SELECT IFNULL(SUM(cs.word_count),0) FROM comments_search cs WHERE cs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ns.word_count),0) FROM site_news_search ns WHERE ns.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(bs.word_count),0) FROM blogs_search bs WHERE bs.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ls.word_count),0) FROM linker_search ls WHERE ls.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ws.word_count),0) FROM writing_search ws WHERE ws.searchwordlist_idsearchwordlist=swl.idsearchwordlist)
       + (SELECT IFNULL(SUM(ips.word_count),0) FROM imagepost_search ips WHERE ips.searchwordlist_idsearchwordlist=swl.idsearchwordlist) AS count
FROM searchwordlist swl
WHERE swl.word LIKE CONCAT(?, '%')
ORDER BY swl.word
LIMIT ? OFFSET ?
`

type WordListWithCountsByPrefixParams struct {
	Prefix interface{}
	Limit  int32
	Offset int32
}

type WordListWithCountsByPrefixRow struct {
	Word  sql.NullString
	Count int32
}

func (q *Queries) WordListWithCountsByPrefix(ctx context.Context, arg WordListWithCountsByPrefixParams) ([]*WordListWithCountsByPrefixRow, error) {
	rows, err := q.db.QueryContext(ctx, wordListWithCountsByPrefix, arg.Prefix, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*WordListWithCountsByPrefixRow
	for rows.Next() {
		var i WordListWithCountsByPrefixRow
		if err := rows.Scan(&i.Word, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const writingSearchDelete = `-- name: WritingSearchDelete :exec
DELETE FROM writing_search
WHERE writing_id=?
`

func (q *Queries) WritingSearchDelete(ctx context.Context, writingID int32) error {
	_, err := q.db.ExecContext(ctx, writingSearchDelete, writingID)
	return err
}

const writingSearchFirst = `-- name: WritingSearchFirst :many
SELECT DISTINCT cs.writing_id
FROM writing_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
`

func (q *Queries) WritingSearchFirst(ctx context.Context, word sql.NullString) ([]int32, error) {
	rows, err := q.db.QueryContext(ctx, writingSearchFirst, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var writing_id int32
		if err := rows.Scan(&writing_id); err != nil {
			return nil, err
		}
		items = append(items, writing_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const writingSearchNext = `-- name: WritingSearchNext :many
SELECT DISTINCT cs.writing_id
FROM writing_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.writing_id IN (/*SLICE:ids*/?)
`

type WritingSearchNextParams struct {
	Word sql.NullString
	Ids  []int32
}

func (q *Queries) WritingSearchNext(ctx context.Context, arg WritingSearchNextParams) ([]int32, error) {
	query := writingSearchNext
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Word)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var writing_id int32
		if err := rows.Scan(&writing_id); err != nil {
			return nil, err
		}
		items = append(items, writing_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
