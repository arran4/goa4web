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
			current CategoryFAQs
		)
		for _, row := range faqRows {
			if current.Category == nil || current.Category.Idfaqcategories.Int32 != row.Idfaqcategories.Int32 {
				if current.Category != nil && current.Category.Idfaqcategories.Int32 != 0 {
					result = append(result, &current)
				}
				current = CategoryFAQs{Category: row}
			}
			current.FAQs = append(current.FAQs, &FAQ{
				CategoryID: int(row.Idfaqcategories.Int32),
				Question:   row.Question.String,
				Answer:     row.Answer.String,
			})
		}
		if current.Category != nil && current.Category.Idfaqcategories.Int32 != 0 {
			result = append(result, &current)
		}
		return result, nil
	})
}
