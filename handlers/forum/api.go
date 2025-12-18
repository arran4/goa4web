package forum

import (
	"encoding/json"
	"net/http"
	"strconv"

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
			text = a4code.Substring(c.Text.String, start, end)
		default:
			text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithParagraphQuote())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"text": text})
}
