package news

import (
	"fmt"
	"io"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func PreviewPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // Limit to 1MB

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok {
		// CoreData should be injected by the middleware stack.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	username := "Guest"
	if u := cd.CurrentUserLoaded(); u != nil {
		username = u.Username.String
	}

	data := struct {
		Content  string
		Username string
	}{
		Content:  string(body),
		Username: username,
	}

	// Set headers for partial HTML content
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := NewsPreviewPageTmpl.Handle(w, r, data); err != nil {
		fmt.Printf("Error processing preview: %v\n", err)
	}
}

const NewsPreviewPageTmpl handlers.Page = "news/preview.gohtml"
