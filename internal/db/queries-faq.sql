-- name: GetFAQUnansweredQuestions :many
SELECT *
FROM faq
WHERE faqCategories_idfaqCategories = '0' OR answer IS NULL;

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

