package forum

import (
	"fmt"
	"github.com/arran4/goa4web/a4code/a4code2html"
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

	conv := a4code2html.New()
	conv.SetInput(string(body))

	// Set headers for partial HTML content
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := io.Copy(w, conv.Process()); err != nil {
		fmt.Printf("Error processing preview: %v\n", err)
		http.Error(w, "Error processing preview", http.StatusInternalServerError)
		return
	}
}
