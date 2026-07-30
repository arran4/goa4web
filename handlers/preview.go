package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if fn, ok := cd.Funcs(r)["a4code2html"].(func(string) template.HTML); ok {
			htmlOut := fn(string(body))
			if _, err := fmt.Fprint(w, htmlOut); err != nil {
				fmt.Printf("Error processing preview: %v\n", err)
				http.Error(w, "Error processing preview", http.StatusInternalServerError)
			}
			return
		}
	}

	// Fallback if CoreData or function is missing
	conv := a4code2html.New()
	conv.SetInput(string(body))
	if _, err := io.Copy(w, conv.Process()); err != nil {
		fmt.Printf("Error processing preview: %v\n", err)
		http.Error(w, "Error processing preview", http.StatusInternalServerError)
		return
	}
}
