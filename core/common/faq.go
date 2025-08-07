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
