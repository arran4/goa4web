package goa4web

import (
	"html/template"
	"net/http"

	"github.com/arran4/goa4web/handlers/common"
)

// NewFuncs delegates to handlers/common.NewFuncs.
func NewFuncs(r *http.Request) template.FuncMap {
	return common.NewFuncs(r)
}
