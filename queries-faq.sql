-- name: SelectUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

-- name: AllQuestions :many
SELECT *
FROM faq;

-- name: Rename_category :exec
UPDATE faqCategories
SET name = ?
WHERE idfaqCategories = ?;

-- name: Delete_category :exec
DELETE FROM faqCategories
WHERE idfaqCategories = ?;

-- name: Create_category :exec
INSERT INTO faqCategories (name)
VALUES (?);

-- name: Add_question :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
VALUES (?, ?, ?);

-- name: Modify_faq :exec
UPDATE faq
SET answer = ?, question = ?, faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: Delete_faq :exec
DELETE FROM faq
WHERE idfaq = ?;

-- name: Faq_categories :many
SELECT idfaqCategories, name
FROM faqCategories;

-- name: Show_questions :many
SELECT c.*, f.*
FROM faq f
LEFT JOIN faqCategories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0 AND f.answer IS NOT NULL
ORDER BY c.idfaqCategories;

