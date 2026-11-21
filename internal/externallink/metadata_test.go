package externallink

import (
	"testing"

	"github.com/arran4/goa4web/internal/opengraph"
)

func TestMetadataParamsPreserveOpenGraphFields(t *testing.T) {
	params := MetadataParams(7, &opengraph.Info{Title: "title", Description: "description", Image: "image", Duration: "duration", UploadDate: "date", Author: "author"})
	if params.ID != 7 || params.CardTitle.String != "title" || !params.CardTitle.Valid || params.CardAuthor.String != "author" {
		t.Fatalf("MetadataParams() = %#v", params)
	}
}
