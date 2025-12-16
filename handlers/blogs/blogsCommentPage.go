package blogs

import (
	"net/http"
)

func BlogsCommentPage(w http.ResponseWriter, r *http.Request) {
	t := NewBlogsCommentTask().(*blogsCommentTask)
	t.Get(w, r)
}
