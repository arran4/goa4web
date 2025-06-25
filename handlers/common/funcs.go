package common

import (
	"html/template"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
)

// NewFuncs delegates to core.common.NewFuncs.
func NewFuncs(r *http.Request) template.FuncMap {
	return corecommon.NewFuncs(r)
}
