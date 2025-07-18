package templates

import (
	"html/template"
	"net/http"
)

// TODO remove
func RenderTemplate(w http.ResponseWriter, name string, data interface{}, funcs template.FuncMap) error {
	return GetCompiledSiteTemplates(funcs).ExecuteTemplate(w, name, data)
}
