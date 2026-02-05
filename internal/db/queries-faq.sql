-- name: AdminGetFAQUnansweredQuestions :many
SELECT *
FROM faq
WHERE category_id IS NULL OR answer IS NULL;

-- name: GetFAQAnsweredQuestions :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.narg(user_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT faq.id, faq.category_id, faq.language_id, faq.author_id, faq.answer, faq.question
FROM faq
WHERE answer IS NOT NULL
  AND deleted_at IS NULL
  AND (
      language_id = 0
      OR language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.narg(user_id)
            AND ul.language_id = faq.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.narg(user_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = faq.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.narg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: AdminGetFAQDismissedQuestions :many
SELECT id, category_id, language_id, author_id, answer, question
FROM faq
WHERE deleted_at IS NOT NULL;

-- name: SystemGetFAQQuestions :many
SELECT *
FROM faq;

-- name: AdminRenameFAQCategory :exec
UPDATE faq_categories
SET name = ?, updated_at = NOW()
WHERE id = ?;

-- name: AdminUpdateFAQCategory :exec
UPDATE faq_categories
SET name = ?, parent_category_id = ?, language_id = ?, priority = ?, updated_at = NOW()
WHERE id = ?;

-- name: AdminDeleteFAQCategory :exec
UPDATE faq_categories SET deleted_at = NOW(), updated_at = NOW()
WHERE id = ?;

-- name: AdminCreateFAQCategory :execresult
INSERT INTO faq_categories (name, parent_category_id, language_id, priority) VALUES (?, ?, ?, ?);

-- name: CreateFAQQuestionForWriter :exec
INSERT INTO faq (question, author_id, language_id)
SELECT sqlc.arg(question), sqlc.arg(writer_id), sqlc.narg(language_id)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'faq'
      AND (g.item = 'question' OR g.item IS NULL)
      AND g.action = 'post'
      AND g.active = 1
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
      ))
);

-- name: InsertFAQQuestionForWriter :execresult
INSERT INTO faq (question, answer, category_id, author_id, language_id, priority)
SELECT sqlc.arg(question), sqlc.arg(answer), sqlc.arg(category_id), sqlc.arg(writer_id), sqlc.narg(language_id), sqlc.arg(priority)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'faq'
      AND (g.item = 'question' OR g.item IS NULL)
      AND g.action = 'post'
      AND g.active = 1
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
      ))
);

-- name: AdminUpdateFAQQuestionAnswer :exec
UPDATE faq
SET answer = ?, question = ?, category_id = ?, updated_at = NOW()
WHERE id = ?;

-- name: AdminDeleteFAQ :exec
UPDATE faq SET deleted_at = NOW(), updated_at = NOW()
WHERE id = ?;

-- name: AdminGetFAQCategories :many
SELECT *
FROM faq_categories
WHERE deleted_at IS NULL;

-- name: AdminListFAQCategories :many
SELECT *
FROM faq_categories
WHERE deleted_at IS NULL
ORDER BY parent_category_id, id;

-- name: AdminGetFAQCategory :one
SELECT * FROM faq_categories WHERE id = ?;

-- name: GetAllAnsweredFAQWithFAQCategoriesForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.narg(user_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT c.id AS category_id, c.name, f.id AS faq_id, f.category_id, f.language_id, f.author_id, f.answer, f.question
FROM faq f
LEFT JOIN faq_categories c ON c.id = f.category_id
WHERE c.id IS NOT NULL
  AND f.answer IS NOT NULL
  AND (
      f.language_id = 0
      OR f.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.narg(user_id)
            AND ul.language_id = f.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.narg(user_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = f.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.narg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY c.id, f.id;

-- name: AdminGetFAQCategoriesWithQuestionCount :many
SELECT c.id, c.parent_category_id, c.language_id, c.name, c.priority, c.updated_at, COUNT(f.id) AS QuestionCount
FROM faq_categories c
LEFT JOIN faq f ON f.category_id = c.id
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.parent_category_id, c.language_id, c.name, c.priority, c.updated_at
ORDER BY c.priority DESC, c.name ASC;


-- name: AdminGetFAQByID :one
SELECT * FROM faq WHERE id = ?;

-- name: GetFAQByID :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.narg(user_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT faq.id, faq.category_id, faq.language_id, faq.author_id, faq.answer, faq.question
FROM faq
WHERE faq.id = sqlc.arg(faq_id)
  AND deleted_at IS NULL
  AND (
      language_id = 0
      OR language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.narg(user_id)
            AND ul.language_id = faq.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.narg(user_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = faq.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.narg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: InsertFAQRevisionForUser :exec
INSERT INTO faq_revisions (faq_id, users_idusers, question, answer, timezone)
SELECT sqlc.arg(faq_id), sqlc.arg(users_idusers), sqlc.arg(question), sqlc.arg(answer), sqlc.arg(timezone)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'faq'
      AND (g.item = 'question' OR g.item IS NULL)
      AND g.action = 'post'
      AND g.active = 1
      AND (g.user_id = sqlc.narg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.narg(user_id)
      ))
);

-- name: GetFAQRevisionsForAdmin :many
SELECT * FROM faq_revisions WHERE faq_id = ? ORDER BY id DESC;

-- name: AdminGetFAQCategoryWithQuestionCountByID :one
SELECT c.id, c.parent_category_id, c.language_id, c.name, c.priority, c.updated_at, COUNT(f.id) AS QuestionCount
FROM faq_categories c
LEFT JOIN faq f ON f.category_id = c.id
WHERE c.id = ?
GROUP BY c.id, c.parent_category_id, c.language_id, c.name, c.priority, c.updated_at;

-- name: AdminGetFAQQuestionsByCategory :many
SELECT * FROM faq WHERE category_id = ? ORDER BY priority DESC, id DESC;

-- name: GetFAQQuestionsByCategory :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.narg(user_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT faq.id, faq.category_id, faq.language_id, faq.author_id, faq.answer, faq.question
FROM faq
WHERE faq.category_id = sqlc.arg(category_id)
  AND deleted_at IS NULL
  AND (
      language_id = 0
      OR language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.narg(user_id)
            AND ul.language_id = faq.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.narg(user_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = faq.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.narg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: AdminUpdateFAQPriority :exec
UPDATE faq SET priority = ?, updated_at = NOW() WHERE id = ?;

-- name: AdminUpdateFAQ :exec
UPDATE faq
SET answer = ?, question = ?, category_id = ?, priority = ?, updated_at = NOW()
WHERE id = ?;

-- name: AdminCreateFAQ :execresult
INSERT INTO faq (question, answer, category_id, author_id, language_id, priority)
VALUES (?, ?, ?, ?, ?, ?);

-- name: AdminMoveFAQContent :exec
UPDATE faq SET category_id = sqlc.arg(new_category_id), updated_at = NOW() WHERE category_id = sqlc.arg(old_category_id);

-- name: AdminMoveFAQChildren :exec
UPDATE faq_categories SET parent_category_id = sqlc.arg(new_parent_id), updated_at = NOW() WHERE parent_category_id = sqlc.arg(old_parent_id);

-- name: AdminGetFAQActiveQuestions :many
SELECT *
FROM faq
WHERE answer IS NOT NULL
  AND category_id IS NOT NULL
  AND deleted_at IS NULL;
