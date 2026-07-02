package forum

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"
)

func QuoteApi(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	quoteId, err := strconv.Atoi(vars["commentid"])
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	replyType := r.URL.Query().Get("type")
	var text string
	if c, err := cd.CommentByID(int32(quoteId)); err == nil && c != nil {
		switch replyType {
		case "paragraph":
			text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithParagraphQuote())
		case "full":
			text = a4code.QuoteText(c.Username.String, c.Text.String)
		case "selected":
			start, _ := strconv.Atoi(r.URL.Query().Get("start"))
			end, _ := strconv.Atoi(r.URL.Query().Get("end"))
			sub, _ := a4code.Substring(c.Text.String, start, end)
			if a4code.IsQuoteBlock(sub) {
				text = sub
			} else {
				text = a4code.QuoteText(c.Username.String, sub)
			}
		default:
			text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithParagraphQuote())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"text": text})
}

type quoteSelectionRequest struct {
	Ranges []quoteSelectionRange `json:"ranges"`
}

type quoteSelectionRange struct {
	CommentID int32 `json:"comment_id"`
	Start     int   `json:"start"`
	End       int   `json:"end"`
}

func QuoteSelectionApi(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	var req quoteSelectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid quote selection request", http.StatusBadRequest)
		return
	}

	var out strings.Builder
	currentUsername := ""
	var currentParts []string

	flush := func() {
		if len(currentParts) == 0 {
			return
		}
		if out.Len() > 0 {
			out.WriteString("\n\n")
		}
		text := strings.Join(currentParts, "\n\n")
		if a4code.IsQuoteBlock(strings.TrimSpace(text)) {
			out.WriteString(text)
			if !strings.HasSuffix(text, "\n") {
				out.WriteByte('\n')
			}
		} else {
			out.WriteString(a4code.QuoteText(currentUsername, text, a4code.WithTrimSpace()))
		}
		currentParts = nil
	}

	for _, selected := range req.Ranges {
		if selected.CommentID == 0 || selected.Start < 0 || selected.End < selected.Start {
			http.Error(w, "Invalid quote selection range", http.StatusBadRequest)
			return
		}
		if selected.Start == selected.End {
			continue
		}

		c, err := cd.CommentByID(selected.CommentID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Comment %d not found", selected.CommentID), http.StatusNotFound)
			return
		}
		if c == nil {
			continue
		}

		sub, err := a4code.Substring(c.Text.String, selected.Start, selected.End)
		if err != nil {
			http.Error(w, "Invalid quote selection range", http.StatusBadRequest)
			return
		}
		sub = strings.TrimSpace(sub)
		if sub == "" {
			continue
		}

		username := c.Username.String
		if len(currentParts) != 0 && username != currentUsername {
			flush()
		}
		currentUsername = username
		currentParts = append(currentParts, sub)
	}
	flush()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"text": out.String()})
}
