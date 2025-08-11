-- name: AdminGetFAQUnansweredQuestions :many
SELECT *
FROM faq
WHERE faq_category_id IS NULL OR answer IS NULL;

-- name: GetFAQAnsweredQuestions :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT faq.id, faq.faq_category_id, faq.language_idlanguage, faq.users_idusers, faq.answer, faq.question
FROM faq
WHERE answer IS NOT NULL
  AND deleted_at IS NULL
  AND (
      language_idlanguage = 0
      OR language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = faq.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = faq.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: AdminGetFAQDismissedQuestions :many
SELECT id, faq_category_id, language_idlanguage, users_idusers, answer, question
FROM faq
WHERE deleted_at IS NOT NULL;

-- name: SystemGetFAQQuestions :many
SELECT *
FROM faq;

-- name: AdminRenameFAQCategory :exec
UPDATE faq_categories
SET name = ?
WHERE id = ?;

-- name: AdminDeleteFAQCategory :exec
UPDATE faq_categories SET deleted_at = NOW()
WHERE id = ?;

-- name: AdminCreateFAQCategory :exec
INSERT INTO faq_categories (name) VALUES (sqlc.arg(name));

-- name: CreateFAQQuestionForWriter :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
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
INSERT INTO faq (question, answer, faq_category_id, users_idusers, language_idlanguage)
SELECT sqlc.arg(question), sqlc.arg(answer), sqlc.arg(category_id), sqlc.arg(writer_id), sqlc.narg(language_id)
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
SET answer = ?, question = ?, faq_category_id = ?
WHERE id = ?;

-- name: AdminDeleteFAQ :exec
UPDATE faq SET deleted_at = NOW()
WHERE id = ?;

-- name: AdminGetFAQCategories :many
SELECT *
FROM faq_categories;

-- name: GetAllAnsweredFAQWithFAQCategoriesForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT c.id AS category_id, c.name, f.id AS faq_id, f.faq_category_id, f.language_idlanguage, f.users_idusers, f.answer, f.question
FROM faq f
LEFT JOIN faq_categories c ON c.id = f.faq_category_id
WHERE c.id IS NOT NULL
  AND f.answer IS NOT NULL
  AND (
      f.language_idlanguage = 0
      OR f.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = f.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = f.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY c.id, f.id;

-- name: AdminGetFAQCategoriesWithQuestionCount :many
SELECT c.*, COUNT(f.id) AS QuestionCount
FROM faq_categories c
LEFT JOIN faq f ON f.faq_category_id = c.id
GROUP BY c.id;


-- name: AdminGetFAQByID :one
SELECT * FROM faq WHERE id = ?;

-- name: GetFAQByID :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT faq.id, faq.faq_category_id, faq.language_idlanguage, faq.users_idusers, faq.answer, faq.question
FROM faq
WHERE faq.id = sqlc.arg(faq_id)
  AND deleted_at IS NULL
  AND (
      language_idlanguage = 0
      OR language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = faq.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = faq.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
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
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
      ))
);

-- name: GetFAQRevisionsForAdmin :many
SELECT * FROM faq_revisions WHERE faq_id = ? ORDER BY id DESC;

-- name: AdminGetFAQCategoryWithQuestionCountByID :one
SELECT c.*, COUNT(f.id) AS QuestionCount
FROM faq_categories c
LEFT JOIN faq f ON f.faq_category_id = c.id
WHERE c.id = ?
GROUP BY c.id;

-- name: AdminGetFAQQuestionsByCategory :many
SELECT * FROM faq WHERE faq_category_id = ?;

-- name: GetFAQQuestionsByCategory :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT faq.id, faq.faq_category_id, faq.language_idlanguage, faq.users_idusers, faq.answer, faq.question
FROM faq
WHERE faq.faq_category_id = sqlc.arg(category_id)
  AND deleted_at IS NULL
  AND (
      language_idlanguage = 0
      OR language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = faq.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND (g.item='question/answer' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.item_id = faq.id OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );
