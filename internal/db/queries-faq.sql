-- name: GetFAQUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

-- name: GetFAQAnsweredQuestions :many
SELECT idfaq, faqCategories_idfaqCategories, language_idlanguage, users_idusers, answer, question
FROM faq
WHERE answer IS NOT NULL AND deleted_at IS NULL;

-- name: GetFAQDismissedQuestions :many
SELECT idfaq, faqCategories_idfaqCategories, language_idlanguage, users_idusers, answer, question
FROM faq
WHERE deleted_at IS NOT NULL;

-- name: GetAllFAQQuestions :many
SELECT *
FROM faq;

-- name: RenameFAQCategory :exec
UPDATE faq_categories
SET name = ?
WHERE idfaqCategories = ?;

-- name: DeleteFAQCategory :exec
UPDATE faq_categories SET deleted_at = NOW()
WHERE idfaqCategories = ?;

-- name: CreateFAQCategory :exec
INSERT INTO faq_categories (name)
VALUES (?);

-- name: CreateFAQQuestion :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
VALUES (?, ?, ?);

-- name: UpdateFAQQuestionAnswer :exec
UPDATE faq
SET answer = ?, question = ?, faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: DeleteFAQ :exec
UPDATE faq SET deleted_at = NOW()
WHERE idfaq = ?;

-- name: GetAllFAQCategories :many
SELECT *
FROM faq_categories;

-- name: GetAllAnsweredFAQWithFAQCategories :many
SELECT c.*, f.*
FROM faq f
LEFT JOIN faq_categories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0 AND f.answer IS NOT NULL
ORDER BY c.idfaqCategories;

-- name: GetFAQCategoriesWithQuestionCount :many
SELECT c.*, COUNT(f.idfaq) AS QuestionCount
FROM faq_categories c
LEFT JOIN faq f ON f.faqCategories_idfaqCategories = c.idfaqCategories
GROUP BY c.idfaqCategories;


-- name: GetFAQByID :one
SELECT * FROM faq WHERE idfaq = ?;

-- renamed to InsertFAQRevisionForUser
-- name: InsertFAQRevisionForUser :exec
INSERT INTO faq_revisions (faq_id, users_idusers, question, answer)
SELECT sqlc.arg(faq_id), sqlc.arg(users_idusers), sqlc.arg(question), sqlc.arg(answer)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'faq'
      AND g.item = 'question'
      AND g.action = 'post'
      AND g.active = 1
      AND g.item_id = sqlc.arg(item_id)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
      ))
);

-- name: AdminInsertFAQRevision :exec
INSERT INTO faq_revisions (faq_id, users_idusers, question, answer)
VALUES (?, ?, ?, ?);

-- name: GetFAQRevisionsForFAQ :many
SELECT * FROM faq_revisions WHERE faq_id = ? ORDER BY id DESC;
