// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries-news.sql

package db

import (
	"context"
	"database/sql"
	"strings"
)

const assignNewsThisThreadId = `-- name: AssignNewsThisThreadId :exec
UPDATE site_news SET forumthread_id = ? WHERE idsiteNews = ?
`

type AssignNewsThisThreadIdParams struct {
	ForumthreadID int32
	Idsitenews    int32
}

func (q *Queries) AssignNewsThisThreadId(ctx context.Context, arg AssignNewsThisThreadIdParams) error {
	_, err := q.db.ExecContext(ctx, assignNewsThisThreadId, arg.ForumthreadID, arg.Idsitenews)
	return err
}

const createNewsPost = `-- name: CreateNewsPost :exec
INSERT INTO site_news (news, users_idusers, occurred, language_idlanguage)
VALUES (?, ?, NOW(), ?)
`

type CreateNewsPostParams struct {
	News               sql.NullString
	UsersIdusers       int32
	LanguageIdlanguage int32
}

func (q *Queries) CreateNewsPost(ctx context.Context, arg CreateNewsPostParams) error {
	_, err := q.db.ExecContext(ctx, createNewsPost, arg.News, arg.UsersIdusers, arg.LanguageIdlanguage)
	return err
}

const deactivateNewsPost = `-- name: DeactivateNewsPost :exec
UPDATE site_news SET deleted_at = NOW() WHERE idsiteNews = ?
`

func (q *Queries) DeactivateNewsPost(ctx context.Context, idsitenews int32) error {
	_, err := q.db.ExecContext(ctx, deactivateNewsPost, idsitenews)
	return err
}

const getForumThreadIdByNewsPostId = `-- name: GetForumThreadIdByNewsPostId :one
SELECT s.forumthread_id, u.idusers
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?
`

type GetForumThreadIdByNewsPostIdRow struct {
	ForumthreadID int32
	Idusers       sql.NullInt32
}

func (q *Queries) GetForumThreadIdByNewsPostId(ctx context.Context, idsitenews int32) (*GetForumThreadIdByNewsPostIdRow, error) {
	row := q.db.QueryRowContext(ctx, getForumThreadIdByNewsPostId, idsitenews)
	var i GetForumThreadIdByNewsPostIdRow
	err := row.Scan(&i.ForumthreadID, &i.Idusers)
	return &i, err
}

const getNewsPostByIdWithWriterIdAndThreadCommentCount = `-- name: GetNewsPostByIdWithWriterIdAndThreadCommentCount :one
SELECT u.username AS writerName, u.idusers as writerId, s.idsitenews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.idsiteNews = ?
`

type GetNewsPostByIdWithWriterIdAndThreadCommentCountRow struct {
	Writername         sql.NullString
	Writerid           sql.NullInt32
	Idsitenews         int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	UsersIdusers       int32
	News               sql.NullString
	Occurred           sql.NullTime
	Comments           sql.NullInt32
}

func (q *Queries) GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx context.Context, idsitenews int32) (*GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
	row := q.db.QueryRowContext(ctx, getNewsPostByIdWithWriterIdAndThreadCommentCount, idsitenews)
	var i GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
	err := row.Scan(
		&i.Writername,
		&i.Writerid,
		&i.Idsitenews,
		&i.ForumthreadID,
		&i.LanguageIdlanguage,
		&i.UsersIdusers,
		&i.News,
		&i.Occurred,
		&i.Comments,
	)
	return &i, err
}

const getNewsPostsByIdsWithWriterIdAndThreadCommentCount = `-- name: GetNewsPostsByIdsWithWriterIdAndThreadCommentCount :many
SELECT u.username AS writerName, u.idusers as writerId, s.idsitenews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.Idsitenews IN (/*SLICE:newsids*/?)
`

type GetNewsPostsByIdsWithWriterIdAndThreadCommentCountRow struct {
	Writername         sql.NullString
	Writerid           sql.NullInt32
	Idsitenews         int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	UsersIdusers       int32
	News               sql.NullString
	Occurred           sql.NullTime
	Comments           sql.NullInt32
}

func (q *Queries) GetNewsPostsByIdsWithWriterIdAndThreadCommentCount(ctx context.Context, newsids []int32) ([]*GetNewsPostsByIdsWithWriterIdAndThreadCommentCountRow, error) {
	query := getNewsPostsByIdsWithWriterIdAndThreadCommentCount
	var queryParams []interface{}
	if len(newsids) > 0 {
		for _, v := range newsids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:newsids*/?", strings.Repeat(",?", len(newsids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:newsids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetNewsPostsByIdsWithWriterIdAndThreadCommentCountRow
	for rows.Next() {
		var i GetNewsPostsByIdsWithWriterIdAndThreadCommentCountRow
		if err := rows.Scan(
			&i.Writername,
			&i.Writerid,
			&i.Idsitenews,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.UsersIdusers,
			&i.News,
			&i.Occurred,
			&i.Comments,
		); err != nil {
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

const getNewsPostsWithWriterUsernameAndThreadCommentCountDescending = `-- name: GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many
SELECT u.username AS writerName, u.idusers as writerId, s.idsitenews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
ORDER BY s.occurred DESC
LIMIT ? OFFSET ?
`

type GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams struct {
	Limit  int32
	Offset int32
}

type GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow struct {
	Writername         sql.NullString
	Writerid           sql.NullInt32
	Idsitenews         int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	UsersIdusers       int32
	News               sql.NullString
	Occurred           sql.NullTime
	Comments           sql.NullInt32
}

func (q *Queries) GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(ctx context.Context, arg GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams) ([]*GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	rows, err := q.db.QueryContext(ctx, getNewsPostsWithWriterUsernameAndThreadCommentCountDescending, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	for rows.Next() {
		var i GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
		if err := rows.Scan(
			&i.Writername,
			&i.Writerid,
			&i.Idsitenews,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.UsersIdusers,
			&i.News,
			&i.Occurred,
			&i.Comments,
		); err != nil {
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

const updateNewsPost = `-- name: UpdateNewsPost :exec
UPDATE site_news SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?
`

type UpdateNewsPostParams struct {
	News               sql.NullString
	LanguageIdlanguage int32
	Idsitenews         int32
}

func (q *Queries) UpdateNewsPost(ctx context.Context, arg UpdateNewsPostParams) error {
	_, err := q.db.ExecContext(ctx, updateNewsPost, arg.News, arg.LanguageIdlanguage, arg.Idsitenews)
	return err
}
