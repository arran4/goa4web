package common

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestAllAnsweredFAQ_Categories(t *testing.T) {
	// Setup
	qs := &db.QuerierStub{}
	cd := NewTestCoreData(t, qs)

	// Mock Data
	category1 := &db.GetAllAnsweredFAQWithFAQCategoriesForUserRow{
		CategoryID: sql.NullInt32{Int32: 1, Valid: true},
		Name:       sql.NullString{String: "Category 1", Valid: true},
		FaqID:      2,
		Question:   sql.NullString{String: "Q1", Valid: true},
		Answer:     sql.NullString{String: "A1", Valid: true},
	}
	category2 := &db.GetAllAnsweredFAQWithFAQCategoriesForUserRow{
		CategoryID: sql.NullInt32{Int32: 3, Valid: true},
		Name:       sql.NullString{String: "Category 2", Valid: true},
		FaqID:      4,
		Question:   sql.NullString{String: "Q2", Valid: true},
		Answer:     sql.NullString{String: "A2", Valid: true},
	}

	qs.GetAllAnsweredFAQWithFAQCategoriesForUserReturns = []*db.GetAllAnsweredFAQWithFAQCategoriesForUserRow{
		category1,
		category2,
	}

	// Execute
	result, err := cd.AllAnsweredFAQ()
	assert.NoError(t, err)

	// Assert
	// Expect 2 categories
	if assert.Len(t, result, 2) {
		// Due to the bug, these assertions might fail if result[0] points to the last category (Category 2)
		assert.Equal(t, int(category1.CategoryID.Int32), int(result[0].Category.CategoryID.Int32), "First category should match Category 1")
		assert.Equal(t, int(category2.CategoryID.Int32), int(result[1].Category.CategoryID.Int32), "Second category should match Category 2")

		assert.Equal(t, "Category 1", result[0].Category.Name.String)
		assert.Equal(t, "Category 2", result[1].Category.Name.String)
	}
}
