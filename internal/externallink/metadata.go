package externallink

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
)

func metadataParams(id int32, info *opengraph.Info) db.UpdateExternalLinkMetadataParams {
	if info == nil {
		info = &opengraph.Info{}
	}
	return db.UpdateExternalLinkMetadataParams{
		ID:              id,
		CardTitle:       sql.NullString{String: info.Title, Valid: info.Title != ""},
		CardDescription: sql.NullString{String: info.Description, Valid: info.Description != ""},
		CardImage:       sql.NullString{String: info.Image, Valid: info.Image != ""},
		CardDuration:    sql.NullString{String: info.Duration, Valid: info.Duration != ""},
		CardUploadDate:  sql.NullString{String: info.UploadDate, Valid: info.UploadDate != ""},
		CardAuthor:      sql.NullString{String: info.Author, Valid: info.Author != ""},
	}
}
