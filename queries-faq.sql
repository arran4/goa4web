-- name: GetFAQUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

-- name: GetAllFAQQuestions :many
SELECT *
FROM faq;

-- name: RenameFAQCategory :exec
UPDATE faqCategories
SET name = ?
WHERE idfaqCategories = ?;

-- name: DeleteFAQCategory :exec
DELETE FROM faqCategories
WHERE idfaqCategories = ?;

-- name: CreateFAQCategory :exec
INSERT INTO faqCategories (name)
VALUES (?);

-- name: CreateFAQQuestion :exec
INSERT INTO faq (question, users_idusers, language_idlanguage)
VALUES (?, ?, ?);

-- name: UpdateFAQQuestionAnswer :exec
UPDATE faq
SET answer = ?, question = ?, faqCategories_idfaqCategories = ?
WHERE idfaq = ?;

-- name: DeleteFAQ :exec
DELETE FROM faq
WHERE idfaq = ?;

-- name: GetAllFAQCategories :many
SELECT *
FROM faqCategories;

-- name: GetAllAnsweredFAQWithFAQCategories :many
SELECT c.*, f.*
FROM faq f
LEFT JOIN faqCategories c ON c.idfaqCategories = f.faqCategories_idfaqCategories
WHERE c.idfaqCategories <> 0 AND f.answer IS NOT NULL
ORDER BY c.idfaqCategories;

