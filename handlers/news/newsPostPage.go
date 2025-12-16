package news

import (
	"net/http"
)

func NewsPostPageHandler(w http.ResponseWriter, r *http.Request) {
	t := NewNewsPostTask().(*newsPostTask)
	t.Get(w, r)
}
