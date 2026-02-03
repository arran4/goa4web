package common

import (
	"database/sql"
	"errors"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

// FAQ represents a single FAQ entry grouped by category.
type FAQ struct {
	CategoryID int
	Question   string
	Answer     string
}

// CategoryFAQs groups FAQ entries under a single category.
type CategoryFAQs struct {
	Category *db.GetAllAnsweredFAQWithFAQCategoriesForUserRow
	FAQs     []*FAQ
}

// AllAnsweredFAQ returns answered FAQ entries grouped by category for the
// current user. Results are cached for the lifetime of the CoreData instance.
func (cd *CoreData) AllAnsweredFAQ() ([]*CategoryFAQs, error) {
	return cd.allAnsweredFAQ.Load(func() ([]*CategoryFAQs, error) {
		if cd.queries == nil {
			return nil, nil
		}
		faqRows, err := cd.queries.GetAllAnsweredFAQWithFAQCategoriesForUser(cd.ctx, db.GetAllAnsweredFAQWithFAQCategoriesForUserParams{
			ViewerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			log.Printf("getAllAnsweredFAQWithFAQCategories Error: %s", err)
			return nil, err
		}
		var (
			result  []*CategoryFAQs
			current *CategoryFAQs
		)
		for _, row := range faqRows {
			if current == nil || current.Category.CategoryID.Int32 != row.CategoryID.Int32 {
				if current != nil && current.Category.CategoryID.Int32 != 0 {
					result = append(result, current)
				}
				current = &CategoryFAQs{Category: row}
			}
			current.FAQs = append(current.FAQs, &FAQ{
				CategoryID: int(row.CategoryID.Int32),
				Question:   row.Question.String,
				Answer:     row.Answer.String,
			})
		}
		if current != nil && current.Category.CategoryID.Int32 != 0 {
			result = append(result, current)
		}
		return result, nil
	})
}

// RenameFAQCategory updates the name of a FAQ category.
func (cd *CoreData) RenameFAQCategory(id int32, name string) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.AdminRenameFAQCategory(cd.ctx, db.AdminRenameFAQCategoryParams{
		Name: sql.NullString{String: name, Valid: true},
		ID:   id,
	})
}

// CreateFAQQuestionParams groups input for CreateFAQQuestion.
type CreateFAQQuestionParams struct {
	Question   string
	Answer     string
	CategoryID int32
	WriterID   int32
	LanguageID int32
}

// CreateFAQQuestion creates a FAQ question and its initial revision.
func (cd *CoreData) CreateFAQQuestion(p CreateFAQQuestionParams) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	res, err := cd.queries.InsertFAQQuestionForWriter(cd.ctx, db.InsertFAQQuestionForWriterParams{
		Question:   sql.NullString{String: p.Question, Valid: p.Question != ""},
		Answer:     sql.NullString{String: p.Answer, Valid: p.Answer != ""},
		CategoryID: sql.NullInt32{Int32: p.CategoryID, Valid: p.CategoryID != 0},
		WriterID:   p.WriterID,
		LanguageID: sql.NullInt32{Int32: p.LanguageID, Valid: p.LanguageID != 0},
		GranteeID:  sql.NullInt32{Int32: p.WriterID, Valid: p.WriterID != 0},
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := cd.queries.InsertFAQRevisionForUser(cd.ctx, db.InsertFAQRevisionForUserParams{
		FaqID:        int32(id),
		UsersIdusers: p.WriterID,
		Question:     sql.NullString{String: p.Question, Valid: p.Question != ""},
		Answer:       sql.NullString{String: p.Answer, Valid: p.Answer != ""},
		Timezone:     sql.NullString{String: cd.Location().String(), Valid: true},
		UserID:       sql.NullInt32{Int32: p.WriterID, Valid: p.WriterID != 0},
		ViewerID:     p.WriterID,
	}); err != nil {
		return 0, err
	}
	return id, nil
}
