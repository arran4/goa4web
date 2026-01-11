package linker

import (
	"net/http"
)

func LinkerCategoryPage(w http.ResponseWriter, r *http.Request) {
	t := NewLinkerCategoryTask().(*linkerCategoryTask)
	t.Get(w, r)
}
