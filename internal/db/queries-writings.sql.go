// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries-writings.sql

package db

import (
	"context"
	"database/sql"
	"strings"
)

const assignWritingThisThreadId = `-- name: AssignWritingThisThreadId :exec
UPDATE writing SET forumthread_id = ? WHERE idwriting = ?
`

type AssignWritingThisThreadIdParams struct {
	ForumthreadID int32
	Idwriting     int32
}

func (q *Queries) AssignWritingThisThreadId(ctx context.Context, arg AssignWritingThisThreadIdParams) error {
	_, err := q.db.ExecContext(ctx, assignWritingThisThreadId, arg.ForumthreadID, arg.Idwriting)
	return err
}

const fetchAllCategories = `-- name: FetchAllCategories :many
SELECT wc.idwritingcategory, wc.writing_category_id, wc.title, wc.description
FROM writing_category wc
`

func (q *Queries) FetchAllCategories(ctx context.Context) ([]*WritingCategory, error) {
	rows, err := q.db.QueryContext(ctx, fetchAllCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*WritingCategory
	for rows.Next() {
		var i WritingCategory
		if err := rows.Scan(
			&i.Idwritingcategory,
			&i.WritingCategoryID,
			&i.Title,
			&i.Description,
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

const fetchCategoriesForUser = `-- name: FetchCategoriesForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT wc.idwritingcategory, wc.writing_category_id, wc.title, wc.description
FROM writing_category wc
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='category'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = wc.idwritingcategory
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
`

type FetchCategoriesForUserParams struct {
	ViewerID int32
	UserID   sql.NullInt32
}

func (q *Queries) FetchCategoriesForUser(ctx context.Context, arg FetchCategoriesForUserParams) ([]*WritingCategory, error) {
	rows, err := q.db.QueryContext(ctx, fetchCategoriesForUser, arg.ViewerID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*WritingCategory
	for rows.Next() {
		var i WritingCategory
		if err := rows.Scan(
			&i.Idwritingcategory,
			&i.WritingCategoryID,
			&i.Title,
			&i.Description,
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

const getAllWritingCategories = `-- name: GetAllWritingCategories :many
SELECT idwritingcategory, writing_category_id, title, description
FROM writing_category
WHERE writing_category_id = ?
`

func (q *Queries) GetAllWritingCategories(ctx context.Context, writingCategoryID int32) ([]*WritingCategory, error) {
	rows, err := q.db.QueryContext(ctx, getAllWritingCategories, writingCategoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*WritingCategory
	for rows.Next() {
		var i WritingCategory
		if err := rows.Scan(
			&i.Idwritingcategory,
			&i.WritingCategoryID,
			&i.Title,
			&i.Description,
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

const getAllWritingsByUser = `-- name: GetAllWritingsByUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = ?
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
`

type GetAllWritingsByUserParams struct {
	ViewerID      int32
	AuthorID      int32
	ViewerMatchID sql.NullInt32
}

type GetAllWritingsByUserRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Username           sql.NullString
	Comments           int64
}

func (q *Queries) GetAllWritingsByUser(ctx context.Context, arg GetAllWritingsByUserParams) ([]*GetAllWritingsByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllWritingsByUser, arg.ViewerID, arg.AuthorID, arg.ViewerMatchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetAllWritingsByUserRow
	for rows.Next() {
		var i GetAllWritingsByUserRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
			&i.Username,
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

const getAllWritingsForIndex = `-- name: GetAllWritingsForIndex :many
SELECT idwriting, title, abstract, writing FROM writing WHERE deleted_at IS NULL
`

type GetAllWritingsForIndexRow struct {
	Idwriting int32
	Title     sql.NullString
	Abstract  sql.NullString
	Writing   sql.NullString
}

func (q *Queries) GetAllWritingsForIndex(ctx context.Context) ([]*GetAllWritingsForIndexRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllWritingsForIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetAllWritingsForIndexRow
	for rows.Next() {
		var i GetAllWritingsForIndexRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.Title,
			&i.Abstract,
			&i.Writing,
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

const getPublicWritings = `-- name: GetPublicWritings :many
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index
FROM writing w
WHERE w.private = 0
ORDER BY w.published DESC
LIMIT ? OFFSET ?
`

type GetPublicWritingsParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) GetPublicWritings(ctx context.Context, arg GetPublicWritingsParams) ([]*Writing, error) {
	rows, err := q.db.QueryContext(ctx, getPublicWritings, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Writing
	for rows.Next() {
		var i Writing
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
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

const getPublicWritingsByUser = `-- name: GetPublicWritingsByUser :many
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = ?
ORDER BY w.published DESC
LIMIT ? OFFSET ?
`

type GetPublicWritingsByUserParams struct {
	UsersIdusers int32
	Limit        int32
	Offset       int32
}

type GetPublicWritingsByUserRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Username           sql.NullString
	Comments           int64
}

func (q *Queries) GetPublicWritingsByUser(ctx context.Context, arg GetPublicWritingsByUserParams) ([]*GetPublicWritingsByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getPublicWritingsByUser, arg.UsersIdusers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetPublicWritingsByUserRow
	for rows.Next() {
		var i GetPublicWritingsByUserRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
			&i.Username,
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

const getPublicWritingsByUserForViewer = `-- name: GetPublicWritingsByUserForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = ?
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?
`

type GetPublicWritingsByUserForViewerParams struct {
	ViewerID int32
	AuthorID int32
	UserID   sql.NullInt32
	Limit    int32
	Offset   int32
}

type GetPublicWritingsByUserForViewerRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Username           sql.NullString
	Comments           int64
}

func (q *Queries) GetPublicWritingsByUserForViewer(ctx context.Context, arg GetPublicWritingsByUserForViewerParams) ([]*GetPublicWritingsByUserForViewerRow, error) {
	rows, err := q.db.QueryContext(ctx, getPublicWritingsByUserForViewer,
		arg.ViewerID,
		arg.AuthorID,
		arg.UserID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetPublicWritingsByUserForViewerRow
	for rows.Next() {
		var i GetPublicWritingsByUserForViewerRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
			&i.Username,
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

const getPublicWritingsInCategory = `-- name: GetPublicWritingsInCategory :many
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writing_category_id=?
ORDER BY w.published DESC
LIMIT ? OFFSET ?
`

type GetPublicWritingsInCategoryParams struct {
	WritingCategoryID int32
	Limit             int32
	Offset            int32
}

type GetPublicWritingsInCategoryRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Username           sql.NullString
	Comments           int64
}

func (q *Queries) GetPublicWritingsInCategory(ctx context.Context, arg GetPublicWritingsInCategoryParams) ([]*GetPublicWritingsInCategoryRow, error) {
	rows, err := q.db.QueryContext(ctx, getPublicWritingsInCategory, arg.WritingCategoryID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetPublicWritingsInCategoryRow
	for rows.Next() {
		var i GetPublicWritingsInCategoryRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
			&i.Username,
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

const getPublicWritingsInCategoryForUser = `-- name: GetPublicWritingsInCategoryForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writing_category_id = ?
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?
`

type GetPublicWritingsInCategoryForUserParams struct {
	ViewerID          int32
	WritingCategoryID int32
	UserID            sql.NullInt32
	Limit             int32
	Offset            int32
}

type GetPublicWritingsInCategoryForUserRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Username           sql.NullString
	Comments           int64
}

func (q *Queries) GetPublicWritingsInCategoryForUser(ctx context.Context, arg GetPublicWritingsInCategoryForUserParams) ([]*GetPublicWritingsInCategoryForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getPublicWritingsInCategoryForUser,
		arg.ViewerID,
		arg.WritingCategoryID,
		arg.UserID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetPublicWritingsInCategoryForUserRow
	for rows.Next() {
		var i GetPublicWritingsInCategoryForUserRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
			&i.Username,
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

const getWritingByIdForUserDescendingByPublishedDate = `-- name: GetWritingByIdForUserDescendingByPublishedDate :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting = ?
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
`

type GetWritingByIdForUserDescendingByPublishedDateParams struct {
	ViewerID      int32
	Idwriting     int32
	ViewerMatchID sql.NullInt32
}

type GetWritingByIdForUserDescendingByPublishedDateRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Writerid           int32
	Writerusername     sql.NullString
}

func (q *Queries) GetWritingByIdForUserDescendingByPublishedDate(ctx context.Context, arg GetWritingByIdForUserDescendingByPublishedDateParams) (*GetWritingByIdForUserDescendingByPublishedDateRow, error) {
	row := q.db.QueryRowContext(ctx, getWritingByIdForUserDescendingByPublishedDate, arg.ViewerID, arg.Idwriting, arg.ViewerMatchID)
	var i GetWritingByIdForUserDescendingByPublishedDateRow
	err := row.Scan(
		&i.Idwriting,
		&i.UsersIdusers,
		&i.ForumthreadID,
		&i.LanguageIdlanguage,
		&i.WritingCategoryID,
		&i.Title,
		&i.Published,
		&i.Writing,
		&i.Abstract,
		&i.Private,
		&i.DeletedAt,
		&i.LastIndex,
		&i.Writerid,
		&i.Writerusername,
	)
	return &i, err
}

const getWritingsByIdsForUserDescendingByPublishedDate = `-- name: GetWritingsByIdsForUserDescendingByPublishedDate :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_idlanguage, w.writing_category_id, w.title, w.published, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting IN (/*SLICE:writing_ids*/?)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
`

type GetWritingsByIdsForUserDescendingByPublishedDateParams struct {
	ViewerID      int32
	WritingIds    []int32
	ViewerMatchID sql.NullInt32
}

type GetWritingsByIdsForUserDescendingByPublishedDateRow struct {
	Idwriting          int32
	UsersIdusers       int32
	ForumthreadID      int32
	LanguageIdlanguage int32
	WritingCategoryID  int32
	Title              sql.NullString
	Published          sql.NullTime
	Writing            sql.NullString
	Abstract           sql.NullString
	Private            sql.NullBool
	DeletedAt          sql.NullTime
	LastIndex          sql.NullTime
	Writerid           int32
	Writerusername     sql.NullString
}

func (q *Queries) GetWritingsByIdsForUserDescendingByPublishedDate(ctx context.Context, arg GetWritingsByIdsForUserDescendingByPublishedDateParams) ([]*GetWritingsByIdsForUserDescendingByPublishedDateRow, error) {
	query := getWritingsByIdsForUserDescendingByPublishedDate
	var queryParams []interface{}
	queryParams = append(queryParams, arg.ViewerID)
	if len(arg.WritingIds) > 0 {
		for _, v := range arg.WritingIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:writing_ids*/?", strings.Repeat(",?", len(arg.WritingIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:writing_ids*/?", "NULL", 1)
	}
	queryParams = append(queryParams, arg.ViewerMatchID)
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetWritingsByIdsForUserDescendingByPublishedDateRow
	for rows.Next() {
		var i GetWritingsByIdsForUserDescendingByPublishedDateRow
		if err := rows.Scan(
			&i.Idwriting,
			&i.UsersIdusers,
			&i.ForumthreadID,
			&i.LanguageIdlanguage,
			&i.WritingCategoryID,
			&i.Title,
			&i.Published,
			&i.Writing,
			&i.Abstract,
			&i.Private,
			&i.DeletedAt,
			&i.LastIndex,
			&i.Writerid,
			&i.Writerusername,
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

const insertWriting = `-- name: InsertWriting :execlastid
INSERT INTO writing (writing_category_id, title, abstract, writing, private, language_idlanguage, published, users_idusers)
VALUES (?, ?, ?, ?, ?, ?, NOW(), ?)
`

type InsertWritingParams struct {
	WritingCategoryID  int32
	Title              sql.NullString
	Abstract           sql.NullString
	Writing            sql.NullString
	Private            sql.NullBool
	LanguageIdlanguage int32
	UsersIdusers       int32
}

func (q *Queries) InsertWriting(ctx context.Context, arg InsertWritingParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertWriting,
		arg.WritingCategoryID,
		arg.Title,
		arg.Abstract,
		arg.Writing,
		arg.Private,
		arg.LanguageIdlanguage,
		arg.UsersIdusers,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const insertWritingCategory = `-- name: InsertWritingCategory :exec
INSERT INTO writing_category (writing_category_id, title, description)
VALUES (?, ?, ?)
`

type InsertWritingCategoryParams struct {
	WritingCategoryID int32
	Title             sql.NullString
	Description       sql.NullString
}

func (q *Queries) InsertWritingCategory(ctx context.Context, arg InsertWritingCategoryParams) error {
	_, err := q.db.ExecContext(ctx, insertWritingCategory, arg.WritingCategoryID, arg.Title, arg.Description)
	return err
}

const listWritersForViewer = `-- name: ListWritersForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username, COUNT(w.idwriting) AS count
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (
    NOT EXISTS (SELECT 1 FROM user_language ul WHERE ul.users_idusers = ?)
    OR w.language_idlanguage IN (
        SELECT ul.language_idlanguage FROM user_language ul WHERE ul.users_idusers = ?
    )
)
AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'writing'
      AND g.item = 'article'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?
`

type ListWritersForViewerParams struct {
	ViewerID int32
	UserID   sql.NullInt32
	Limit    int32
	Offset   int32
}

type ListWritersForViewerRow struct {
	Username sql.NullString
	Count    int64
}

func (q *Queries) ListWritersForViewer(ctx context.Context, arg ListWritersForViewerParams) ([]*ListWritersForViewerRow, error) {
	rows, err := q.db.QueryContext(ctx, listWritersForViewer,
		arg.ViewerID,
		arg.ViewerID,
		arg.ViewerID,
		arg.UserID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ListWritersForViewerRow
	for rows.Next() {
		var i ListWritersForViewerRow
		if err := rows.Scan(&i.Username, &i.Count); err != nil {
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

const searchWritersForViewer = `-- name: SearchWritersForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username, COUNT(w.idwriting) AS count
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (LOWER(u.username) LIKE LOWER(?) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(?))
  AND (
    NOT EXISTS (SELECT 1 FROM user_language ul WHERE ul.users_idusers = ?)
    OR w.language_idlanguage IN (
        SELECT ul.language_idlanguage FROM user_language ul WHERE ul.users_idusers = ?
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'writing'
      AND g.item = 'article'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = w.idwriting
      AND (g.user_id = ? OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?
`

type SearchWritersForViewerParams struct {
	ViewerID int32
	Query    string
	UserID   sql.NullInt32
	Limit    int32
	Offset   int32
}

type SearchWritersForViewerRow struct {
	Username sql.NullString
	Count    int64
}

func (q *Queries) SearchWritersForViewer(ctx context.Context, arg SearchWritersForViewerParams) ([]*SearchWritersForViewerRow, error) {
	rows, err := q.db.QueryContext(ctx, searchWritersForViewer,
		arg.ViewerID,
		arg.Query,
		arg.Query,
		arg.ViewerID,
		arg.ViewerID,
		arg.UserID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SearchWritersForViewerRow
	for rows.Next() {
		var i SearchWritersForViewerRow
		if err := rows.Scan(&i.Username, &i.Count); err != nil {
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

const setWritingLastIndex = `-- name: SetWritingLastIndex :exec
UPDATE writing SET last_index = NOW() WHERE idwriting = ?
`

func (q *Queries) SetWritingLastIndex(ctx context.Context, idwriting int32) error {
	_, err := q.db.ExecContext(ctx, setWritingLastIndex, idwriting)
	return err
}

const updateWriting = `-- name: UpdateWriting :exec
UPDATE writing
SET title = ?, abstract = ?, writing = ?, private = ?, language_idlanguage = ?
WHERE idwriting = ?
`

type UpdateWritingParams struct {
	Title              sql.NullString
	Abstract           sql.NullString
	Writing            sql.NullString
	Private            sql.NullBool
	LanguageIdlanguage int32
	Idwriting          int32
}

func (q *Queries) UpdateWriting(ctx context.Context, arg UpdateWritingParams) error {
	_, err := q.db.ExecContext(ctx, updateWriting,
		arg.Title,
		arg.Abstract,
		arg.Writing,
		arg.Private,
		arg.LanguageIdlanguage,
		arg.Idwriting,
	)
	return err
}

const updateWritingCategory = `-- name: UpdateWritingCategory :exec
UPDATE writing_category
SET title = ?, description = ?, writing_category_id = ?
WHERE idwritingCategory = ?
`

type UpdateWritingCategoryParams struct {
	Title             sql.NullString
	Description       sql.NullString
	WritingCategoryID int32
	Idwritingcategory int32
}

func (q *Queries) UpdateWritingCategory(ctx context.Context, arg UpdateWritingCategoryParams) error {
	_, err := q.db.ExecContext(ctx, updateWritingCategory,
		arg.Title,
		arg.Description,
		arg.WritingCategoryID,
		arg.Idwritingcategory,
	)
	return err
}
