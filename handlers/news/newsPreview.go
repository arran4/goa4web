package news

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"io"
	"net/http"
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

	if err := cd.ExecuteSiteTemplate(w, r, "news/preview.gohtml", data); err != nil {
		fmt.Printf("Error processing preview: %v\n", err)
		http.Error(w, "Error processing preview", http.StatusInternalServerError)
		return
	}
}
