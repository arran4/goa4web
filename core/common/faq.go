package common

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// RenameFAQCategory updates the name of a FAQ category.
func (cd *CoreData) RenameFAQCategory(id int32, name string) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.AdminRenameFAQCategory(cd.ctx, db.AdminRenameFAQCategoryParams{
		Name:            sql.NullString{String: name, Valid: true},
		Idfaqcategories: id,
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
		CategoryID: p.CategoryID,
		WriterID:   p.WriterID,
		LanguageID: p.LanguageID,
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
		UserID:       sql.NullInt32{Int32: p.WriterID, Valid: p.WriterID != 0},
		ViewerID:     p.WriterID,
	}); err != nil {
		return 0, err
	}
	return id, nil
}

