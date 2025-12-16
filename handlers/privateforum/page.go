package privateforum

import (
	"net/http"
)

func PrivateForumPage(w http.ResponseWriter, r *http.Request) {
	t := NewPrivateForumTask().(*privateForumTask)
	t.Get(w, r)
}
