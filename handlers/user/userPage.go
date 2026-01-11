package user

import (
	"net/http"
)

func UserPage(w http.ResponseWriter, r *http.Request) {
	t := NewUserTask().(*userTask)
	t.Get(w, r)
}
