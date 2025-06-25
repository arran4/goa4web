package goa4web

import (
	"html/template"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
)

// NewFuncs delegates to common.NewFuncs.
func NewFuncs(r *http.Request) template.FuncMap {
	return common.NewFuncs(r)
}
