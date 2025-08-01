-- name: AdminGetFAQUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

-- name: AdminGetFAQAnsweredQuestions :many
SELECT idfaq, faqCategories_idfaqCategories, language_idlanguage, users_idusers, answer, question
FROM faq
WHERE answer IS NOT NULL AND deleted_at IS NULL;

-- name: AdminGetFAQDismissedQuestions :many
SELECT idfaq, faqCategories_idfaqCategories, language_idlanguage, users_idusers, answer, question
FROM faq
WHERE deleted_at IS NOT NULL;

-- name: GetAllFAQQuestions :many
SELECT *
FROM faq;

-- name: GetAllFAQQuestionsForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT f.*
FROM faq f
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='faq'
      AND g.item='question'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = f.idfaq
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
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
  );

-- name: AdminRenameFAQCategory :exec
UPDATE faq_categories
SET name = ?
WHERE idfaqCategories = ?;

-- name: AdminDeleteFAQCategory :exec
UPDATE faq_categories SET deleted_at = NOW()
WHERE idfaqCategories = ?;

-- name: AdminCreateFAQCategory :exec
INSERT INTO faq_categories (name)
VALUES (?);

-- name: CreateFAQQuestion :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
VALUES (?, ?, ?);

-- name: AdminUpdateFAQQuestionAnswer :exec
UPDATE faq
SET answer = ?, question = ?, faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: AdminDeleteFAQ :exec
UPDATE faq SET deleted_at = NOW()
WHERE idfaq = ?;

-- name: AdminGetAllFAQCategories :many
SELECT *
FROM faq_categories;

-- name: GetAllAnsweredFAQWithFAQCategories :many
SELECT c.*, f.*
FROM faq f
LEFT JOIN faq_categories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0 AND f.answer IS NOT NULL
ORDER BY c.idfaqCategories;

-- name: GetAllAnsweredFAQWithFAQCategoriesForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT c.*, f.*
FROM faq f
LEFT JOIN faq_categories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0
  AND f.answer IS NOT NULL
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='faq'
        AND g.item='question'
        AND g.action='see'
        AND g.active=1
        AND g.item_id = f.idfaq
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
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
ORDER BY c.idfaqCategories;

-- name: AdminGetFAQCategoriesWithQuestionCount :many
SELECT c.*, COUNT(f.idfaq) AS QuestionCount
FROM faq_categories c
LEFT JOIN faq f ON f.faqCategories_idfaqCategories = c.idfaqCategories
GROUP BY c.idfaqCategories;


-- name: AdminGetFAQByID :one
SELECT * FROM faq WHERE idfaq = ?;

-- name: InsertFAQRevision :exec
INSERT INTO faq_revisions (faq_id, users_idusers, question, answer)
VALUES (?, ?, ?, ?);

-- name: InsertFAQRevisionForUser :exec
INSERT INTO faq_revisions (faq_id, users_idusers, question, answer)
SELECT sqlc.arg(faq_id), sqlc.arg(users_idusers), sqlc.arg(question), sqlc.arg(answer)
FROM faq f
WHERE f.idfaq = sqlc.arg(faq_id)
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
        AND g.item='question'
        AND g.action='post'
        AND g.active=1
        AND g.item_id = f.idfaq
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
        ))
  );

-- name: AdminGetFAQRevisionsForFAQ :many
SELECT * FROM faq_revisions WHERE faq_id = ? ORDER BY id DESC;
