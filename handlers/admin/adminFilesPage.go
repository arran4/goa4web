package admin

import (
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
)

func AdminFilesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Files"
	type Data struct {
		Path    string
		Parent  string
		Entries []ImageFileEntry
	}

	ttlStr := r.URL.Query().Get("ttl")
	ttl := 24 * time.Hour
	if ttlStr != "" {
		if d, err := time.ParseDuration(ttlStr); err == nil {
			ttl = d
		}
	}

	listing, err := BuildImageFilesListing(r.Context(), cd.Queries(), cd.Config.ImageUploadDir, r.URL.Query().Get("path"), cd.ImageSignKey, cd.SignImageURL, ttl)
	if err != nil {
		var invalidPath invalidPathError
		var notFound notFoundError
		switch {
		case errors.As(err, &invalidPath):
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid path"))
		case errors.As(err, &notFound):
			handlers.RenderErrorPage(w, r, fmt.Errorf("not found"))
		default:
			log.Printf("readdir: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	data := Data{
		Path:    listing.Path,
		Parent:  listing.Parent,
		Entries: listing.Entries,
	}

	AdminFilesPageTmpl.Handle(w, r, data)
}

// AdminFilesPageTmpl renders the admin image files page.
const AdminFilesPageTmpl tasks.Template = "admin/adminFilesPage.gohtml"
