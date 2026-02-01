package admin

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

func AdminUnmanagedFilesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Unmanaged Files"
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

	// Filter for unmanaged files
	var unmanaged []ImageFileEntry
	for _, e := range listing.Entries {
		if e.IsDir || !e.IsManaged {
			unmanaged = append(unmanaged, e)
		}
	}

	data := Data{
		Path:    listing.Path,
		Parent:  listing.Parent,
		Entries: unmanaged,
	}

	AdminUnmanagedFilesPageTmpl.Handle(w, r, data)
}

// AdminUnmanagedFilesPageTmpl renders the admin unmanaged files page.
const AdminUnmanagedFilesPageTmpl tasks.Template = "admin/adminUnmanagedFilesPage.gohtml"
