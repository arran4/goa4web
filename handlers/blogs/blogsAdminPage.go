package blogs

import (
	"net/http"
)

// BlogsAdminPage shows the blog administration index with a list of blogs.
func BlogsAdminPage(w http.ResponseWriter, r *http.Request) {
	t := NewBlogsAdminTask().(*blogsAdminTask)
	t.Get(w, r)
}
